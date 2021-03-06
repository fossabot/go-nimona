package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"nimona.io/go/codec"
	"nimona.io/go/log"
	"nimona.io/go/primitives"
)

func (api *API) HandleGetStreams(c *gin.Context) {
	ns := c.Param("ns")
	pattern := c.Param("pattern")

	if pattern != "" {
		pattern = ns + pattern
	} else {
		pattern = ns
	}

	write := func(conn *websocket.Conn, data interface{}) error {
		contentType := strings.ToLower(c.ContentType())
		if strings.Contains(contentType, "cbor") {
			bs, err := codec.Marshal(data)
			if err != nil {
				return err
			}
			if err := conn.WriteMessage(2, bs); err != nil {
				return err
			}
		}

		return conn.WriteJSON(data)
	}

	var wsupgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	ctx := context.Background()
	logger := log.Logger(ctx).Named("api")
	signer := api.addressBook.GetLocalPeerKey()
	incoming := make(chan *primitives.Block, 100)
	outgoing := make(chan *BlockRequest, 100)

	go func() {
		for {
			select {
			case v := <-incoming:
				m := api.mapBlock(v)
				write(conn, m)

			case r := <-outgoing:
				v := &primitives.Block{
					Type: r.Type,
					Annotations: &primitives.Annotations{
						Policies: []primitives.Policy{
							primitives.Policy{
								Subjects: r.Recipients,
								Actions: []string{
									"read",
								},
								Effect: "allow",
							},
						},
					},
					Payload: r.Payload,
				}
				m := api.mapBlock(v)
				m["status"] = "ok"
				if err := primitives.Sign(v, signer); err != nil {
					logger.Error("could not sign outgoing block", zap.Error(err))
					m["status"] = "error signing block"
					// TODO handle error
					write(conn, m)
					continue
				}
				for _, recipient := range r.Recipients {
					addr := "peer:" + recipient
					if err := api.exchange.Send(ctx, v, addr); err != nil {
						logger.Error("could not send outgoing block", zap.Error(err))
						m["status"] = "error sending block"
					}
					// TODO handle error
					write(conn, m)
				}
			}
		}
	}()
	fmt.Println(pattern, pattern, pattern, pattern, pattern, pattern, pattern)
	hr, err := api.exchange.Handle(pattern, func(v *primitives.Block) error {
		incoming <- v
		return nil
	})
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	defer hr()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if err == io.EOF {
				logger.Debug("ws conn is dead", zap.Error(err))
				return
			}

			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				logger.Debug("ws conn closed", zap.Error(err))
				return
			}

			if websocket.IsUnexpectedCloseError(err) {
				logger.Warn("ws conn closed with unexpected error", zap.Error(err))
				return
			}

			logger.Warn("could not read from ws", zap.Error(err))
			continue
		}
		r := &BlockRequest{}
		if err := json.Unmarshal(msg, r); err != nil {
			logger.Error("could not unmarshal outgoing block", zap.Error(err))
			continue
		}
		if r.Type == "" || len(r.Recipients) == 0 {
			// TODO send error message to ws
			logger.Error("outgoing block missing type or recipients")
			continue
		}
		outgoing <- r
	}
}
