package trademc

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorcon/rcon"
	"github.com/gorcon/rcon/rcontest"
	"log/slog"
	"net"
)

var r *RCON

func R() *RCON {
	if r == nil {
		panic("rcon is not initialized")
	}
	return r
}

type RCON struct {
	server *rcontest.Server
	l      *slog.Logger

	hashedPass string
	passSalt   []byte
}

type Config struct {
	ListenPort int64
	Logger     *slog.Logger
	Password   string
}

func New(c Config) *RCON {
	r = &RCON{}
	r.passSalt = []byte(uuid.NewString() + uuid.NewString() + uuid.NewString())

	if c.ListenPort == 0 {
		c.ListenPort = 13337
	}
	if c.Logger == nil {
		c.Logger = slog.Default()
	}
	r.l = c.Logger

	originalPassword := "password"
	if c.Password != "" {
		originalPassword = c.Password
	}
	r.hashedPass = hashString(originalPassword, r.passSalt)

	r.server = rcontest.NewServer(
		rcontest.SetSettings(rcontest.Settings{
			Password: r.hashedPass,
		}),
		rcontest.SetAuthHandler(r.auth),
		rcontest.SetCommandHandler(r.handleMessage),
	)

	r.server.Listener = r.mustListenAt(int(c.ListenPort))
	return R()
}

func (r *RCON) Close() error {
	defer r.l.Info("Closed RCON server.")
	r.l.Info("Shutting down RCON server...")
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// just prevent panic to make sure that script is still running
			}
		}()
		r.server.Close()
	}()
	return nil
}

func (r *RCON) auth(c *rcontest.Context) {
	msg := c.Request().Body()
	clientPassword := hashString(msg, r.passSalt)

	if clientPassword != r.hashedPass {
		r.responseFail(c)
		return
	}

	r.responseSuccess(c)
}

func hashString(s string, salt []byte) string {
	hash := sha256.New()
	hash.Write(append(salt, []byte(s)...))

	return hex.EncodeToString(hash.Sum(nil))
}

func (r *RCON) responseFail(c *rcontest.Context) {
	_, _ = rcon.NewPacket(rcon.SERVERDATA_AUTH_RESPONSE, -1, string([]byte{0x00})).WriteTo(c.Conn())
}

func (r *RCON) responseSuccess(c *rcontest.Context) {
	_, _ = rcon.NewPacket(rcon.SERVERDATA_AUTH_RESPONSE, c.Request().ID, "").WriteTo(c.Conn())
}

func (r *RCON) responseMessage(msg string) func(*rcontest.Context) {
	return func(c *rcontest.Context) {
		_, _ = rcon.NewPacket(rcon.SERVERDATA_RESPONSE_VALUE, c.Request().ID, msg).WriteTo(c.Conn())
	}
}

func (r *RCON) mustListenAt(port int) net.Listener {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(fmt.Sprintf("rcon: failed to listen on a port: %v", err))
	}

	return l
}
