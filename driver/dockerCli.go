package driver

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"io"
	"os"
	"time"
)

var Cli *client.Client

func initialCli() error {
	var err error
	if Cli == nil {
		Cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			return err
		}
	}
	return err
}

//镜像拉取
func DockerPull(path, authbase string) (string, error) {
	var err error
	initialCli()
	fmt.Println("镜像地址：" + path)
	ctx := context.Background()
	options := types.ImagePullOptions{RegistryAuth: authbase, All: true}
	reader, err := Cli.ImagePull(ctx, path, options)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)
	return buf.String(), err
}

//容器创建
func DockerCreate(path, name, restart string, binds []string, portbinds nat.PortMap, exposedports nat.PortSet, env []string, networkmode string, auth string) (container.ContainerCreateCreatedBody, error) {
	var err error
	DockerPull(path, auth)
	ctx := context.Background()
	config := &container.Config{
		Image:     path,
		Tty:       true,
		OpenStdin: true,
		Env:       env,
	}
	config.ExposedPorts = exposedports
	resp, err := Cli.ContainerCreate(ctx, config, &container.HostConfig{
		Binds:           binds,
		ContainerIDFile: "",
		LogConfig:       container.LogConfig{},
		NetworkMode:     container.NetworkMode(networkmode),
		PortBindings:    portbinds,
		RestartPolicy:   container.RestartPolicy{restart, 0},
		AutoRemove:      false,
		VolumeDriver:    "",
		VolumesFrom:     nil,
		CapAdd:          nil,
		CapDrop:         nil,
		Capabilities:    nil,
		CgroupnsMode:    "",
		DNS:             nil,
		DNSOptions:      nil,
		DNSSearch:       nil,
		ExtraHosts:      nil,
		GroupAdd:        nil,
		IpcMode:         "",
		Cgroup:          "",
		Links:           nil,
		OomScoreAdj:     0,
		PidMode:         "",
		Privileged:      false,
		PublishAllPorts: false,
		ReadonlyRootfs:  false,
		SecurityOpt:     nil,
		StorageOpt:      nil,
		Tmpfs:           nil,
		UTSMode:         "",
		UsernsMode:      "",
		ShmSize:         0,
		Sysctls:         nil,
		Runtime:         "",
		ConsoleSize:     [2]uint{},
		Isolation:       "",
		Resources:       container.Resources{},
		Mounts:          nil,
		MaskedPaths:     nil,
		ReadonlyPaths:   nil,
		Init:            nil,
	}, &network.NetworkingConfig{}, name)
	if err != nil {
		return container.ContainerCreateCreatedBody{}, err
	}
	return resp, err
}

//容器移除
func DockerRemove(id string) error {
	var err error
	initialCli()
	ctx := context.Background()
	if err = Cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{
		RemoveVolumes: false,
		RemoveLinks:   false,
		Force:         true,
	}); err != nil {
		return err
	}
	return err
}

//镜像移除
func DockerRemoveImage(id string) ([]types.ImageDeleteResponseItem, error) {
	var err error
	initialCli()
	ctx := context.Background()
	ress, err := Cli.ImageRemove(ctx, id, types.ImageRemoveOptions{
		Force:         false,
		PruneChildren: false,
	})
	if err != nil {
		return nil, err
	}
	return ress, err
}

//容器开启
func DockerStart(id string) error {
	var err error
	initialCli()
	ctx := context.Background()
	if err = Cli.ContainerStart(ctx, id, types.ContainerStartOptions{}); err != nil {
		return err
	}
	return nil
}
func DockerRestart(id string) error {
	var err error
	initialCli()
	ctx := context.Background()
	if err = Cli.ContainerRestart(ctx, id, nil); err != nil {
		return err
	}
	return err
}

//本地容器枚举
func DockerPs(all bool) ([]types.Container, error) {
	var err error
	initialCli()
	containers, err := Cli.ContainerList(context.Background(), types.ContainerListOptions{All: all})
	if err != nil {
		return nil, err
	}

	return containers, err
}

//容器停止
func DockerStop(id string) error {
	var err error
	initialCli()
	//containers,err:=dockerPs(false)
	ctx := context.Background()
	t := 10 * time.Second
	if err = Cli.ContainerStop(ctx, id, &t); err != nil {
		return err
	}
	return err
}

//容器日志输出
func DockerLogs(id string) (string, error) {
	var err error
	initialCli()
	ctx := context.Background()
	options := types.ContainerLogsOptions{ShowStdout: true}
	out, err := Cli.ContainerLogs(ctx, id, options)
	if err != nil {
		return "", nil
	}
	defer out.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(out)
	return buf.String(), err
}

//本地镜像枚举
func DockerImages() ([]types.ImageSummary, error) {
	var err error
	initialCli()
	ctx := context.Background()
	images, err := Cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		return nil, err
	}
	return images, err
}

//容器生成镜像(用于快照远程容器环境)
func DockerCommit(id string, reference string) (types.IDResponse, error) {
	var err error
	initialCli()
	ctx := context.Background()
	commitResp, err := Cli.ContainerCommit(ctx, id, types.ContainerCommitOptions{Reference: reference})
	if err != nil {
		return types.IDResponse{}, err
	}
	return commitResp, err
}

//镜像打标签
func DockerTag(id, path string) error {
	var err error
	initialCli()
	ctx := context.Background()
	err = Cli.ImageTag(ctx, id, path)
	if err != nil {
		return err
	}
	return err
}

//镜像推送
func DockerPush(path, authbase string) error {
	var err error
	initialCli()
	ctx := context.Background()
	reader, err := Cli.ImagePush(ctx, path, types.ImagePushOptions{RegistryAuth: authbase})
	if err != nil {
		return err
	}
	defer reader.Close()
	io.Copy(os.Stdout, reader)
	return err
}

//运行容器
func DockerRun(path, name, restart string, binds []string, portbinds nat.PortMap, exportedports nat.PortSet, env []string, networkmode string, auth string) (container.ContainerCreateCreatedBody, error) {
	var err error
	if restart == "" {
		restart = "no"
	}
	resp, err := DockerCreate(path, name, restart, binds, portbinds, exportedports, env, networkmode, auth)
	if err != nil {
		return container.ContainerCreateCreatedBody{}, err
	}
	err = DockerStart(resp.ID)
	if err != nil {
		return container.ContainerCreateCreatedBody{}, err
	}
	return resp, err
}

//镜像检索
func ImageSearch(term, autbase string) ([]registry.SearchResult, error) {
	var err error
	initialCli()
	ctx := context.Background()
	reader, err := Cli.ImageSearch(ctx, term, types.ImageSearchOptions{
		RegistryAuth:  autbase,
		PrivilegeFunc: nil,
		Filters:       filters.Args{},
		Limit:         10,
	})
	return reader, err
}
