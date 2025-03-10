package punishment

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/k4ties/dystopia/plugins/practice/src/user"
	"strings"
)

type User struct {
	Name string
	UUID uuid.UUID
	XUID string

	DeviceIDs []string
	IPs       []string
}

func (u User) string() string {
	return fmt.Sprintf(
		"%s;%s;%s;%s;%s",
		u.Name,
		u.UUID.String(),
		u.XUID,
		strings.Join(u.IPs, ","),
		strings.Join(u.DeviceIDs, ","),
	)
}

func (u User) Equals(u2 User) bool {
	return u.string() == u2.string()
}

func mustUserFromString(s string) User {
	u, err := userFromString(s)
	if err != nil {
		panic("cannot parse user from string: " + err.Error())
	}

	return u
}

func userFromString(s string) (User, error) {
	parts := strings.Split(s, ";")
	if len(parts) != 5 {
		return User{}, errors.New("must be 5 split by ';' parts")
	}

	name := parts[0]
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return User{}, err
	}
	xuid := parts[2]
	ips := strings.Split(parts[3], ",")
	deviceIds := strings.Split(parts[4], ",")

	if len(ips) == 0 {
		return User{}, errors.New("must have at least one IP address")
	}
	if len(deviceIds) == 0 {
		return User{}, errors.New("must have at least one device ID")
	}

	return User{
		Name:      name,
		UUID:      id,
		XUID:      xuid,
		DeviceIDs: deviceIds,
		IPs:       ips,
	}, nil
}

func FromPracticeUser(u *user.User) User {
	return User{
		Name:      u.Data().Name(),
		UUID:      u.Data().UUID(),
		XUID:      u.Data().XUID(),
		DeviceIDs: u.Data().DIDs(),
		IPs:       u.Data().IPs(),
	}
}
