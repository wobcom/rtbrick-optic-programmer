package connection

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/rtbrick"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"gopkg.in/yaml.v3"
)

type RouterConnection struct {
	Router       string
	ClientConfig *ssh.ClientConfig
	SSHAgent     *agent.Agent
	SSHClient    *ssh.Client
	SFTPClient   *sftp.Client
}

func New(user, router string) (*RouterConnection, error) {
	var sshAgent agent.Agent
	agentConn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil, err
	}
	sshAgent = agent.NewClient(agentConn)

	signers, err := sshAgent.Signers()
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signers...)},
	}

	return &RouterConnection{
		ClientConfig: config,
		Router:       router,
		SSHAgent:     &sshAgent,
	}, nil
}

func (r *RouterConnection) Connect() error {

	sshClient, err := ssh.Dial("tcp", net.JoinHostPort(r.Router, "1022"), r.ClientConfig)
	if err != nil {
		return err
	}

	r.SSHClient = sshClient

	sftpClient, err := sftp.NewClient(r.SSHClient)
	if err != nil {
		return err
	}
	r.SFTPClient = sftpClient

	return nil
}

func (r *RouterConnection) Close() error {
	if r.SFTPClient != nil {
		err := r.SFTPClient.Close()
		if err != nil {
			return err
		}
		r.SFTPClient = nil
	}
	if r.SSHClient != nil {
		err := r.SSHClient.Close()
		if err != nil {
			return err
		}
		r.SSHClient = nil
	}
	if r.SSHClient != nil {
		err := r.SSHClient.Close()
		if err != nil {
			return err
		}
		r.SSHClient = nil
	}
	return nil
}

func (r *RouterConnection) GetDeviceInformation() (*rtbrick.ImageMetadata, *rtbrick.PPDConfiguration, error) {
	imageMetadataFile, err := r.SFTPClient.Open("/etc/rtbrick/image/metadata.yaml")
	if err != nil {
		return nil, nil, errors.Errorf("could not open image metadata file, %v", err)
	}

	imageMetadataContent, err := io.ReadAll(imageMetadataFile)
	if err != nil {
		return nil, nil, errors.Errorf("could not read content from image metadata file, %v", err)
	}

	var imageMetadata rtbrick.ImageMetadata

	err = yaml.Unmarshal(imageMetadataContent, &imageMetadata)
	if err != nil {
		return nil, nil, errors.Errorf("could not parse content from image metadata file, %v", err)
	}

	peripheraldConfigPath := fmt.Sprintf("/usr/share/rtbrick/peripherald/config/platforms/%s/%s.json", imageMetadata.Manufacturer, imageMetadata.Model)

	ppdConfigFile, err := r.SFTPClient.Open(peripheraldConfigPath)
	if err != nil {
		return nil, nil, errors.Errorf("could not open peripherald config file, %v", err)
	}

	content, err := io.ReadAll(ppdConfigFile)
	if err != nil {
		return nil, nil, errors.Errorf("could not read content from peripherald config file, %v", err)
	}

	var ppdConfig rtbrick.PPDConfiguration

	err = json.Unmarshal(content, &ppdConfig)
	if err != nil {
		return nil, nil, errors.Errorf("could not parse content from peripherald config file, %v", err)
	}

	return &imageMetadata, &ppdConfig, nil
}

func (r *RouterConnection) RunSSHCommand(command string) (string, error) {
	session, err := r.SSHClient.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var stdoutBuffer, stderrBuffer bytes.Buffer
	session.Stdout = &stdoutBuffer
	session.Stderr = &stderrBuffer

	log.Printf("Running command=%v", command)
	err = session.Run(command)
	if err != nil {
		return "", err
	}

	return stdoutBuffer.String(), nil
}

func (r *RouterConnection) GetI2CDump(i2cbusId int, page int) ([]byte, error) {
	_, err := r.RunSSHCommand(fmt.Sprintf("sudo i2cset -y %d 0x50 127 %d", i2cbusId, page))
	if err != nil {
		return nil, err
	}
	out, err := r.RunSSHCommand(fmt.Sprintf("sudo i2cdump -y %d 0x50 b", i2cbusId))
	if err != nil {
		return nil, err
	}
	_, err = r.RunSSHCommand(fmt.Sprintf("sudo i2cset -y %d 0x50 127 %d", i2cbusId, 0))
	if err != nil {
		return nil, err
	}
	return rtbrick.ParseI2CDump(out)
}

func (r *RouterConnection) DoI2CSet(i2cbusId int, page int, byte int, value byte) error {
	_, err := r.RunSSHCommand(fmt.Sprintf("sudo i2cset -y %d 0x50 127 %d", i2cbusId, page))
	if err != nil {
		return err
	}
	_, err = r.RunSSHCommand(fmt.Sprintf("sudo i2cset -y %d 0x50 %d %d", i2cbusId, byte, value))
	if err != nil {
		return err
	}
	_, err = r.RunSSHCommand(fmt.Sprintf("sudo i2cset -y %d 0x50 127 %d", i2cbusId, 0))
	if err != nil {
		return err
	}
	return nil
}
