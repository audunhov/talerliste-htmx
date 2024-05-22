package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/tursodatabase/go-libsql"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

type Reciever struct {
	*echo.Response
	Id int
}

func newReciever(r *echo.Response, id int) Reciever {
	return Reciever{
		Response: r,
		Id:       id,
	}
}

func main() {

	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbName := os.Getenv("DB_NAME")
	primaryUrl := os.Getenv("DB_URL")
	authToken := os.Getenv("DB_TOKEN")

	dir, err := os.MkdirTemp("", "libsql-*")
	if err != nil {
		fmt.Println("Error creating temporary directory:", err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	dbPath := filepath.Join(dir, dbName)

	connector, err := libsql.NewEmbeddedReplicaConnector(dbPath, primaryUrl,
		libsql.WithAuthToken(authToken),
	)
	if err != nil {
		fmt.Println("Error creating connector:", err)
		os.Exit(1)
	}
	defer connector.Close()

	db := sql.OpenDB(connector)
	defer db.Close()

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	renderer := newTemplate()
	e.Renderer = renderer

	channel := make(chan string)

	e.File("/favicon.ico", "public/favicon.ico")

	page := newPage()

	prows, err := db.Query("SELECT * FROM speakers")
	if err != nil {
		log.Fatal("Could not query")
	}
	defer prows.Close()

	var participants Participants
	var pidmap = make(map[int]Participant)
	for prows.Next() {
		var participant Participant

		if err := prows.Scan(&participant.Id, &participant.Name, &participant.Gender); err != nil {
			fmt.Println("Error scanning row,", err)
		}

		pidmap[participant.Id] = participant
		participants = append(participants, participant)
	}

	if err := prows.Err(); err != nil {
		fmt.Println("Error during rows iteration,", err)
	}

	ttrows, err := db.Query("SELECT * FROM talk_types")
	if err != nil {
		log.Fatal("Could not query")
	}
	defer ttrows.Close()

	var talkTypes TalkTypes
	var ttidmap = make(map[int]TalkType)
	for ttrows.Next() {
		var talkType TalkType

		if err := ttrows.Scan(&talkType.Id, &talkType.Name, &talkType.MaxReplies, &talkType.Color); err != nil {
			fmt.Println("Error scanning row,", err)
		}

		ttidmap[talkType.Id] = talkType
		talkTypes = append(talkTypes, talkType)
	}

	if err := ttrows.Err(); err != nil {
		fmt.Println("Error during rows iteration,", err)
	}

	trows, err := db.Query("SELECT * FROM talks")
	if err != nil {
		log.Fatal("Could not query")
	}
	defer trows.Close()

	var talks Talks
	for trows.Next() {
		var talk Talk

		if err := trows.Scan(&talk.Id, &talk.Participant, &talk.Type); err != nil {
			fmt.Println("Error scanning row,", err)
		}

		participant := pidmap[talk.Participant]
		talk_type := ttidmap[talk.Type]

		page.TalkBlocks = append(page.TalkBlocks, newTalkBlock(talk, participant, talk_type))

		talks = append(talks, talk)
	}

	if err := trows.Err(); err != nil {
		fmt.Println("Error during rows iteration,", err)
	}

	page.TalkTypes = talkTypes
	page.Participants = participants

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index", page)
	})

	e.GET("/talerliste", func(c echo.Context) error {
		talkBlocks := page.TalkBlocks
		return c.Render(http.StatusOK, "talerliste", talkBlocks)
	})

	e.GET("/add-participant", func(c echo.Context) error {
		return c.Render(200, "newParticipant", nil)
	})

	e.POST("/participant/file", func(c echo.Context) error {
		formFile, err := c.FormFile("participants")
		if err != nil {
			return err
		}

		parts := strings.Split(formFile.Filename, ".")
		extension := parts[len(parts)-1]

		switch extension {
		case "csv", "xlsx":
			println("Jippi!")
		default:
			println("Oopsie :(")
		}

		file, err := formFile.Open()
		defer file.Close()

		if err != nil {
			return err
		}

		participantsFromCsv(file)

		return c.String(200, "Yeehaw")
	})

	e.POST("/participant", func(c echo.Context) error {
		name := c.FormValue("name")
		gender := c.FormValue("gender")

		stmt, err := db.Prepare("INSERT INTO speakers (name, gender) VALUES (?, ?)")

		if err != nil {
			return err
		}

		var id int
		err = stmt.QueryRow(name, gender).Scan(&id)
		if err != nil {
			fmt.Println("Could not insert into db", err)
		}

		participant := newParticipant(id, name, gender)
		page.Participants = append(page.Participants, participant)

		c.Render(http.StatusOK, "addParticipant", participant)
		return c.Render(http.StatusOK, "participant", participant)
	})

	recievers := []Reciever{}
	recieverId := 0

	e.GET("/talk", func(c echo.Context) error {

		w := c.Response()
		id := recieverId
		recievers = append(recievers, newReciever(w, id))
		recieverId++

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		r := c.Request()
		_, cancel := context.WithCancel(r.Context())
		defer cancel()
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-r.Context().Done():
					for idx, reciever := range recievers {
						if reciever.Id == id {
							recievers = append(recievers[:idx], recievers[idx+1:]...)
						}
					}

					return
				case <-ticker.C:
					event := Event{
						Comment: []byte("keepalive"),
					}
					event.MarshalTo(w)
					w.Flush()

				}
			}
		}()

		for data := range channel {
			event := Event{
				Event: []byte("add"),
				Data:  []byte(data),
			}

			for _, reciever := range recievers {
				event.MarshalTo(reciever)
				reciever.Flush()
			}
		}

		return nil
	})

	e.POST("/talk", func(c echo.Context) error {
		person, err := strconv.Atoi(c.FormValue("participant"))

		if err != nil {
			return c.String(400, "UHHHH")
		}

		talk_type, err := strconv.Atoi(c.FormValue("type"))

		if err != nil {
			return c.String(400, "UHHHH")
		}

		stmt, err := db.Prepare("INSERT INTO talks (type, speaker) VALUES (?, ?) RETURNING id")

		if err != nil {
			return err
		}
		defer stmt.Close()

		var id int
		err = stmt.QueryRow(talk_type, person).Scan(&id)
		if err != nil {
			fmt.Println("Could not insert into db", err)
		}

		talk := newTalk(id, person, talk_type, nil)

		participant := page.Participants.getParticipantById(person)
		talkType := page.TalkTypes.getTalkTypeById(talk_type)
		block := newTalkBlock(talk, *participant, *talkType)

		page.TalkBlocks = append(page.TalkBlocks, block)

		buf := new(bytes.Buffer)
		renderer.Render(buf, "talk", block, c)
		channel <- buf.String()

		return c.NoContent(http.StatusCreated)
	})

	e.DELETE("/talk/:id", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.String(http.StatusUnprocessableEntity, "Could not parse id")
		}
		var index = -1
		var oldBlock TalkBlock

		for i, v := range page.TalkBlocks {
			if v.Talk.Id == id {
				index = i
				oldBlock = v
				break
			}
		}

		if index == -1 {
			return c.String(http.StatusNotFound, "Could not find talk")
		}

		_, err = db.Exec("DELETE FROM talks WHERE id = ?", id)

		if err != nil {
			fmt.Println("Could not delete from db", err)
		}

		page.TalkBlocks = append(page.TalkBlocks[:index], page.TalkBlocks[index+1:]...)

		buf := new(bytes.Buffer)
		renderer.Render(buf, "deleteTalk", oldBlock, c)
		channel <- buf.String()

		return c.NoContent(http.StatusNoContent)
	})

	e.Logger.Fatal(e.Start(":42069"))

}

func participantsFromCsv(r io.Reader) error {

	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	data, err := reader.ReadAll()
	if err != nil {
		return err
	}

	for _, row := range data {
		for _, col := range row {
			print("%s, ", col)
		}
		print("\n")
	}

	return nil

}
