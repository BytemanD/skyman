package guest

import (
	"fmt"
	"strings"
	"time"

	"github.com/BytemanD/go-console/console"
	"libvirt.org/go/libvirt"
)

type Guest struct {
	Connection string
	Domain     string
	ByUUID     bool
	QGATimeout int
	conn       libvirt.Connect
	domain     *libvirt.Domain
	domainName string
}

type FioOptions struct {
	RW       string
	Numjobs  int
	Runtime  int
	Size     string
	FileName string
	IODepth  int
	BS       string
}

func (guest *Guest) Connect() error {
	console.Debug("connecting to %s ...", guest)
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
	guest.domain = domain
	guest.domainName, _ = domain.GetName()
	return nil
}
func (guest *Guest) getDoamin() *libvirt.Domain {
	if guest.domain == nil {
		guest.Connect()
	}
	return guest.domain
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
	return fmt.Sprintf("<Guest %s @%s>", g.Domain, g.Connection)
}
func (g Guest) IsRunning() bool {
	domainInfo, err := g.domain.GetInfo()
	if err != nil {
		return false
	}
	return domainInfo.State == libvirt.DOMAIN_RUNNING
}
func (g Guest) IsShutoff() bool {
	domainInfo, err := g.domain.GetInfo()
	if err != nil {
		return false
	}
	return domainInfo.State == libvirt.DOMAIN_SHUTOFF
}

func (g Guest) RunFio(opts FioOptions) (string, error) {
	if err := g.Connect(); err != nil {
		return "", fmt.Errorf("连接实例失败, %s", err)
	}
	rmFile := false
	if opts.FileName == "" {
		console.Warn("disk path is none, use root path")
		opts.FileName = fmt.Sprintf("/iotest_%s", time.Now().Format("2006-01-02_150405"))
		rmFile = true
	}
	fioCmd := []string{
		"fio", fmt.Sprintf("-name=Fio_%s_Test", opts.RW),
		"-group_reporting", "-ioengine=libaio", "-direct=1",
		fmt.Sprintf("-rw=%s", opts.RW),
		fmt.Sprintf("-size=%s", opts.Size),
		fmt.Sprintf("-numjobs=%d", opts.Numjobs),
		fmt.Sprintf("-runtime=%d", opts.Runtime),
		fmt.Sprintf("-iodepth=%d", opts.IODepth),
		fmt.Sprintf("-bs=%s", opts.BS),
		fmt.Sprintf("-filename=%s", opts.FileName),
	}
	fmt.Printf(">> %s\n", strings.Join(fioCmd, " "))
	execResult := g.Exec(strings.Join(fioCmd, " "), true)
	if execResult.Failed {
		console.Error("test failed: %s\n%s", execResult.OutData, execResult.ErrData)
		return "", fmt.Errorf("run fio failed")
	}
	if rmFile {
		g.Exec("rm -rf "+opts.FileName, true)
	}
	return execResult.OutData, nil
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
