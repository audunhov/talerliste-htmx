package main

import (
	"html/template"
	"io"
	"net/http"
	"strconv"

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

var participantId = 0

type Participant struct {
	Id   int
	Name string
}

func newParticipant(name string) Participant {
	participantId++
	return Participant{
		Id:   participantId,
		Name: name,
	}
}

type Participants []Participant

func (p *Participants) getParticipantById(id int) *Participant {
	for _, p := range *p {
		if p.Id == id {
			return &p
		}
	}
	return nil
}

func newParticipants() Participants {
	return Participants{
		newParticipant("Audun"),
		newParticipant("Mali"),
	}
}

var talkId = 0

type Talk struct {
	Id          int
	Type        string
	Participant int
}

func newTalk(participantId int) Talk {
	talkId++
	return Talk{
		Id:          talkId,
		Type:        "innlegg",
		Participant: participantId,
	}
}

type Storage interface {
	savePage() error
	loadPage() Page
}

type Talks []Talk

func newTalks() Talks {
	return Talks{}
}

type Page struct {
	Participants Participants
	Talks        Talks
}

func newPage() Page {
	return Page{
		Participants: newParticipants(),
		Talks:        newTalks(),
	}
}

type Message struct {
	message string
}

func newMessage(message string) Message {
	return Message{
		message: message,
	}
}

type TalkBlock struct {
  Talk Talk
  Participant Participant
}

func newTalkBlock(talk Talk, participant Participant) TalkBlock {
  return TalkBlock{
    Talk: talk,
    Participant: participant,
  }
}

func main() {

	e := echo.New()

	e.Use(middleware.Logger())

	e.Renderer = newTemplate()

	page := newPage()

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index", page)
	})

	e.POST("/talk/:id", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.String(http.StatusBadRequest, "Could not parse id")
		}
		talk := newTalk(id)

    participant := page.Participants.getParticipantById(id)
    block := newTalkBlock(talk, *participant)

		page.Talks = append(page.Talks, talk)

		return c.Render(http.StatusOK, "talk", block)
	})


	e.DELETE("/talk/:id", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.String(http.StatusUnprocessableEntity, "Could not parse id")
		}
		var index = -1
		for i, v := range page.Talks {
			if v.Id == id {
				index = i
				break
			}
		}

		if index == -1 {
			return c.String(http.StatusNotFound, "Could not find talk")
		}

		page.Talks = append(page.Talks[:index], page.Talks[index+1:]...)

		return c.NoContent(http.StatusOK)
	})

	e.Logger.Fatal(e.Start(":42069"))

}
