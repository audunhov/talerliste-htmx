package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	renderer := newTemplate()
	e.Renderer = renderer

	channel := make(chan string)

	e.File("/favicon.ico", "public/favicon.ico")

	page := newPage()

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index", page)
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
		participant := newParticipant(name, gender)
		page.Participants = append(page.Participants, participant)
		c.Render(http.StatusOK, "genderOption", gender)
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

		talk := newTalk(person, talk_type, nil)

		participant := page.Participants.getParticipantById(person)
		talkType := page.TalkTypes.getTalkTypeById(talk_type)
		block := newTalkBlock(talk, *participant, *talkType)

		page.TalkBlocks = append(page.TalkBlocks, block)

		buf := new(bytes.Buffer)
		renderer.Render(buf, "talk", block, c)
		channel <- buf.String()

		return c.NoContent(200)
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

		page.TalkBlocks = append(page.TalkBlocks[:index], page.TalkBlocks[index+1:]...)

		buf := new(bytes.Buffer)
		renderer.Render(buf, "deleteTalk", oldBlock, c)
		channel <- buf.String()

		return c.NoContent(http.StatusOK)
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
