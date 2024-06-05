package main

import (
	"audunhov/talerliste/cmd/types"
	"audunhov/talerliste/views"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"path/filepath"
	"strconv"
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

// func getDbs() (*TursoDBs, error) {
// 	org := "audunhov"
// 	url := fmt.Sprintf("https://api.turso.tech/v1/organizations/%s/databases", org)
//
// 	client := &http.Client{}
// 	req, err := http.NewRequest("GET", url, nil)
//
// 	authHeader := fmt.Sprintf("Bearer %s", os.Getenv("API_TOKEN"))
// 	req.Header.Set("Authorization", authHeader)
//
// 	resp, err := client.Do(req)
//
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	body, err := io.ReadAll(resp.Body)
//
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var Dbs TursoDBs
// 	err = json.Unmarshal(body, &Dbs)
//
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return &Dbs, nil
// }

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

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

	channel := make(chan string)

	e.File("/favicon.ico", "public/favicon.ico")

	page := types.NewPage()

	prows, err := db.Query("SELECT * FROM speakers")
	if err != nil {
		log.Fatal("Could not query")
	}
	defer prows.Close()

	var participants types.Participants
	var pidmap = make(map[int]types.Participant)
	for prows.Next() {
		var participant types.Participant

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

	var talkTypes types.TalkTypes
	var ttidmap = make(map[int]types.TalkType)
	for ttrows.Next() {
		var talkType types.TalkType

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

	var talks types.Talks
	for trows.Next() {
		var talk types.Talk

		if err := trows.Scan(&talk.Id, &talk.Participant, &talk.Type); err != nil {
			fmt.Println("Error scanning row,", err)
		}

		participant := pidmap[talk.Participant]
		talk_type := ttidmap[talk.Type]

		page.TalkBlocks = append(page.TalkBlocks, types.NewTalkBlock(talk, participant, talk_type))

		talks = append(talks, talk)
	}

	if err := trows.Err(); err != nil {
		fmt.Println("Error during rows iteration,", err)
	}

	page.TalkTypes = talkTypes
	page.Participants = participants

	e.GET("/", func(c echo.Context) error {
		return views.HomePage(page).Render(c.Request().Context(), c.Response().Writer)
	})

	e.GET("/talerliste", func(c echo.Context) error {
		return views.Talerliste(page).Render(c.Request().Context(), c.Response().Writer)
	})

	e.GET("/add-participant", func(c echo.Context) error {
		return views.NewParticipant().Render(c.Request().Context(), c.Response().Writer)
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

		participant := types.NewParticipant(id, name, gender)
		page.Participants = append(page.Participants, participant)

		views.AddParticipant(participant).Render(c.Request().Context(), c.Response().Writer)
		return views.ParticipantDisp(participant).Render(c.Request().Context(), c.Response().Writer)
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
					event := types.Event{
						Comment: []byte("keepalive"),
					}
					event.MarshalTo(w)
					w.Flush()

				}
			}
		}()

		for data := range channel {
			event := types.Event{
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

		talk := types.NewTalk(id, person, talk_type, nil)

		var participant types.Participant

		err = db.QueryRow("SELECT * FROM speakers WHERE id = ?", person).Scan(&participant.Id, &participant.Name, &participant.Gender)

		if err != nil {
			return err
		}

		talkType := page.TalkTypes.GetTalkTypeById(talk_type)
		block := types.NewTalkBlock(talk, participant, *talkType)

		page.TalkBlocks = append(page.TalkBlocks, block)

		buf := new(bytes.Buffer)
		views.Talk(block, true).Render(c.Request().Context(), buf)
		channel <- buf.String()

		return c.NoContent(http.StatusCreated)
	})

	e.DELETE("/talk/:id", onDeleteTalk(page, channel, db))

	e.Logger.Fatal(e.Start(":42069"))

}

func onDeleteTalk(page types.Page, channel chan string, db *sql.DB) func(c echo.Context) error {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.String(http.StatusUnprocessableEntity, "Could not parse id")
		}
		var index = -1
		var oldBlock types.TalkBlock

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

		go func() {
			_, err = db.Exec("DELETE FROM talks WHERE id = ?", id)

			if err != nil {
				fmt.Println("Could not delete from db", err)
			}
		}()

		page.TalkBlocks = append(page.TalkBlocks[:index], page.TalkBlocks[index+1:]...)

		buf := new(bytes.Buffer)
		views.DeleteTalk(oldBlock).Render(c.Request().Context(), buf)
		channel <- buf.String()

		return c.NoContent(http.StatusNoContent)
	}
}
