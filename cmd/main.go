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
func main() {

	e := echo.New()

	e.Use(middleware.Logger())
	e.Renderer = newTemplate()

	e.File("/favicon.ico", "public/favicon.ico")

	page := newPage()

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index", page)
	})

	e.POST("/participant", func(c echo.Context) error {
		name := c.FormValue("name")
		participant := newParticipant(name)
		page.Participants = append(page.Participants, participant)
		return c.Render(http.StatusOK, "participant", participant)
	})

	e.POST("/talk/:id", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.String(http.StatusBadRequest, "Could not parse id")
		}
		talk := newTalk(id)

		participant := page.Participants.getParticipantById(id)
		block := newTalkBlock(talk, *participant)

		page.TalkBlocks = append(page.TalkBlocks, block)

		return c.Render(http.StatusOK, "talk", block)
	})

	e.DELETE("/talk/:id", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.String(http.StatusUnprocessableEntity, "Could not parse id")
		}
		var index = -1
		for i, v := range page.TalkBlocks {
			if v.Talk.Id == id {
				index = i
				break
			}
		}

		if index == -1 {
			return c.String(http.StatusNotFound, "Could not find talk")
		}

		page.TalkBlocks = append(page.TalkBlocks[:index], page.TalkBlocks[index+1:]...)

		return c.NoContent(http.StatusOK)
	})

	e.Logger.Fatal(e.Start(":42069"))

}
