package guest

import (
	"fmt"
	"strings"

	"libvirt.org/go/libvirt"
)

type Guest struct {
	Connection string
	Domain     string
	ByUUID     bool
	QGATimeout int
	conn       libvirt.Connect
	domain     libvirt.Domain
	domainName string
}

func (guest *Guest) Connect() error {
	conn, err := libvirt.NewConnect(fmt.Sprintf("qemu+tcp://%s/system", guest.Connection))
	if err != nil {
		return err
	}
	guest.conn = *conn
	var (
		domain *libvirt.Domain
	)
	domain, err = conn.LookupDomainByUUIDString(guest.Domain)
	if err != nil {
		domain, err = conn.LookupDomainByName(guest.Domain)
	}
	if err != nil {
		return err
	}
	guest.domain = *domain
	guest.domainName, _ = domain.GetName()
	return nil
}

func (g Guest) IsSame(other Guest) bool {
	return g.Connection == other.Connection && g.Domain == other.Domain
}
func (g *Guest) GetName() string {
	if g.domainName == "" {
		g.domainName, _ = g.domain.GetName()
	}
	return g.domainName
}

func (g Guest) String() string {
	return fmt.Sprintf("<%s %s>", g.Connection, g.Domain)
}
func (g Guest) IsRunning() bool {
	domainInfo, err := g.domain.GetInfo()
	if err != nil {
		return false
	}
	return domainInfo.State == libvirt.DOMAIN_RUNNING
}

func ParseGuest(guestConnector string) (*Guest, error) {
	values := strings.Split(guestConnector, ":")
	var connection, domain string
	switch len(values) {
	case 1:
		connection, domain = "localhost", values[0]
	case 2:
		connection, domain = values[0], values[1]
	default:
		return nil, fmt.Errorf("invlid guest connection: %s", guestConnector)
	}

	return &Guest{
		Connection: connection,
		Domain:     domain,
	}, nil
}
