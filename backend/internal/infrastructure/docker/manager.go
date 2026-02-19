package docker

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

type ContainerManager struct {
	cli *client.Client
}

func NewManager(host string) (*ContainerManager, error) {
	cli, err := client.NewClientWithOpts(client.WithHost(host), client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &ContainerManager{cli: cli}, nil
}

func (m *ContainerManager) CreateContainer(ctx context.Context, img string, cmd []string) (string, error) {
	resp, err := m.cli.ContainerCreate(ctx, &container.Config{
		Image:        img,
		Cmd:          cmd,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		OpenStdin:    true,
		Tty:          false,
	}, nil, nil, nil, "")
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (m *ContainerManager) StartContainer(ctx context.Context, id string) error {
	return m.cli.ContainerStart(ctx, id, container.StartOptions{})
}

func (m *ContainerManager) Attach(ctx context.Context, id string) (types.HijackedResponse, error) {
	return m.cli.ContainerAttach(ctx, id, container.AttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
	})
}

func (m *ContainerManager) StopContainer(ctx context.Context, id string) error {
	timeout := 5 
	return m.cli.ContainerStop(ctx, id, container.StopOptions{Timeout: &timeout})
}

func (m *ContainerManager) RemoveContainer(ctx context.Context, id string) error {
	return m.cli.ContainerRemove(ctx, id, container.RemoveOptions{Force: true})
}

func (m *ContainerManager) PullImage(ctx context.Context, img string) (io.ReadCloser, error) {
	return m.cli.ImagePull(ctx, img, image.PullOptions{})
}