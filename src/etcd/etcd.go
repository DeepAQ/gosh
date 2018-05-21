package etcd

import (
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
	"os"
	"net"
	"context"
)

func Register(etcd string, port int) error {
	fmt.Printf("Registering to etcd %s\n", etcd)
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcd},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create etcd client:", err)
		return err
	}
	addresses, _ := net.InterfaceAddrs()
	ip := ""
	for _, addr := range addresses {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.IsGlobalUnicast() && ipnet.IP.To4() != nil {
			ip = ipnet.IP.String()
		}
	}
	if ip == "" {
		fmt.Fprintln(os.Stderr, "Failed to get IP address")
		return err
	}
	ipport := fmt.Sprintf("%s:%d", ip, port)
	lease := clientv3.NewLease(cli)
	grant, err := lease.Grant(context.TODO(), 10)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to grant lease from etcd:", err)
		return err
	}
	_, err = cli.Put(context.TODO(), "dubbomesh/"+ipport, ipport, clientv3.WithLease(grant.ID))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to put to etcd:", err)
		return err
	}
	fmt.Println("Register success:", ipport)
	go func() {
		for {
			time.Sleep(5 * time.Second)
			//fmt.Println("Heartbeat ...")
			_, err = lease.KeepAliveOnce(context.TODO(), grant.ID)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Failed to send heartbeat:", err)
			}
		}
	}()
	return nil
}

func Query(etcd string) ([][]byte, error) {
	fmt.Printf("Querying from etcd %s\n", etcd)
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcd},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create etcd client:", err)
		return nil, err
	}
	resp, err := cli.Get(context.TODO(), "dubbomesh/", clientv3.WithPrefix())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to query etcd:", err)
		return nil, err
	}
	fmt.Println("Providers:", resp.Kvs)
	servers := make([][]byte, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		servers[i] = kv.Value
	}
	return servers, nil
}
