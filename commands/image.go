package commands

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/slack-go/slack"
	"google.golang.org/api/customsearch/v1"
)

type SubHandlerImage struct {
	c *slack.Client
	s *customsearch.Service
}

func NewSubHandlerImage(c *slack.Client, s *customsearch.Service) *SubHandlerImage {
	return &SubHandlerImage{c, s}
}

func (o *SubHandlerImage) Name() string {
	return "image"
}

func (o *SubHandlerImage) Handle(slashCmd *slack.SlashCommand) error {
	q := strings.Join(strings.Fields(slashCmd.Text)[1:], "")
	log.Println(q)
	resp, err := o.s.Cse.List().SearchType("image").Cx("83bd9114c4919450d").Q(q).Start(1).Do()
	if err != nil {
		return err
	}
	if len(resp.Items) == 0 {
		return errors.New("the length of resp.Items must be greater than 0")
	}
	for _, item := range resp.Items {
		log.Println(item.Link)
	}

	hresp, err := http.Get(resp.Items[0].Link)
	if err != nil {
		return err
	}
	defer hresp.Body.Close()

	fileParam := slack.FileUploadParameters{
		Reader:   hresp.Body,
		Filename: q,
		Channels: []string{slashCmd.ChannelID},
	}
	if _, err := o.c.UploadFile(fileParam); err != nil {
		return err
	}

	return err
}
