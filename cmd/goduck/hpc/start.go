package hpc

import (
	"fmt"

	"github.com/codeskyblue/go-sh"
)

const RED = "\033[0;31m"
const BLUE = "\033[0;34m"
const NC = "\033[0m"

func Start(mode string, ips []string, username, fpath, fname, cpath, lpath, lname string) error {
	ip_map := make(map[string]int)
	for idx, ip := range ips {
		_, ok := ip_map[ip]
		if !ok {
			ip_map[ip] = idx + 1
		}
	}

	err := uploadFile(ips, username, fpath, fname, cpath, lpath, lname, ip_map)
	if err != nil {
		return err
	}

	err = modifyConfig(mode, ips, username, fpath, fname, cpath, lpath, lname)
	if err != nil {
		return err
	}

	err = startNode(ips, username, fpath, fname, cpath, lpath, lname)
	if err != nil {
		return err
	}

	err = startMq(ips, username, fpath, fname, cpath, lpath, lname, ip_map)
	if err != nil {
		return err
	}

	return nil
}

func startMq(ips []string, username string, fpath string, fname string, cpath string, lpath string, lname string, ip_map map[string]int) error {
	for idx, ip := range ips {
		fmt.Printf("%s====> Starting mq server in node%d%s\n", BLUE, idx+1, NC)
		v, _ := ip_map[ip]
		if v == idx+1 {
			who := fmt.Sprintf("%s@%s", username, ip)
			err := sh.
				Command("ssh", who, fmt.Sprintf("docker run -d --name %srabbitmq -p 5672:5672 -p 15672:15672 docker.io/rabbitmq:3-management", username)).
				Run()
			if err != nil {
				return err
			}
		} else {
			fmt.Printf("Node%d is on the same server as node%d, so the mq server has been started.\n", idx+1, v)
		}
	}
	return nil
}

func startNode(ips []string, username string, fpath string, fname string, cpath string, lpath string, lname string) error {
	for idx, ip := range ips {
		fmt.Printf("%s====> Starting hyperchain node%d%s\n", BLUE, idx+1, NC)
		who := fmt.Sprintf("%s@%s", username, ip)
		err := sh.Command("ssh", who, fmt.Sprintf("cd ~/node%d && ./start.sh", idx+1)).
			Run()
		if err != nil {
			return err
		}
	}
	return nil
}

func modifyConfig(mode string, ips []string, username string, fpath string, fname string, cpath string, lpath string, lname string) error {
	for idx, ip := range ips {
		fmt.Printf("%s====> Modifying the configuration of node%d%s\n", BLUE, idx+1, NC)
		who := fmt.Sprintf("%s@%s", username, ip)
		err := sh.Command("ssh", who, fmt.Sprintf(
			"sed -i 's/\"domain1 127.0.0.1:50011\",/\"domain1 %s:5001%d\",/g' ~/node%d/addr.toml "+
				"&& sed -i '3c id          = %d' ~/node%d/namespaces/global/config/peerconfig.toml "+
				"&& sed -i '4c hostname    = \"node%d\"' ~/node%d/namespaces/global/config/peerconfig.toml "+
				"&& sed -i 's/8081/808%d/g' ~/node%d/global.toml "+
				"&& sed -i 's/9001/900%d/g' ~/node%d/global.toml "+
				"&& sed -i 's/10001/1000%d/g' ~/node%d/global.toml "+
				"&& sed -i 's/50081/5008%d/g' ~/node%d/global.toml "+
				"&& sed -i 's/50051/5005%d/g' ~/node%d/global.toml "+
				"&& sed -i 's/50011/5001%d/g' ~/node%d/global.toml",
			ip, idx+1, idx+1,
			idx+1, idx+1,
			idx+1, idx+1,
			idx+1, idx+1,
			idx+1, idx+1,
			idx+1, idx+1,
			idx+1, idx+1,
			idx+1, idx+1,
			idx+1, idx+1)).
			Run()
		if err != nil {
			return err
		}
	}

	if mode == "solo" {
		who := fmt.Sprintf("%s@%s", username, ips[0])
		err := sh.Command("ssh", who, fmt.Sprintf(
			"sed -i 's/algo = \"RBFT\"/algo = \"SOLO\"/g' ~/node1/namespaces/global/config/namespace.toml "+
				"&& sed -i 's/batch_size       = 500/batch_size       = 20/g' ~/node1/namespaces/global/config/namespace.toml "+
				"&& sed -i 's/pool_size        = 50000/pool_size        = 2000/g' ~/node1/namespaces/global/config/namespace.toml "+
				"&& n1=`sed -n '/consensus.solo.timeout/=' ~/node1/namespaces/global/config/namespace.toml` "+
				"&& sed -i ''$n1'a set	= \"0.1s\"' ~/node1/namespaces/global/config/namespace.toml "+
				"&& sed -i ''$n1'i slice_size	= 25' ~/node1/namespaces/global/config/namespace.toml "+
				"&& sed -i '$a self = \"node1\"' ~/node1/addr.toml "+
				"&& sed -i '1,$d' ~/node1/hosts.toml "+
				"&& sed -i '2,$d' ~/node1/namespaces/global/config/peerconfig.toml "+
				"&& sed -i '1a n\\t= 1\\nhostname\\t= \\\"node1\\\"\\nnew\\t= false\\nvp\\t= true\\ncaconf\\t= \\\"config/namespace.toml\\\"\\n[[nodes]]\\nostname\\t= \\\"node1\\\"' ~/node1/namespaces/global/config/peerconfig.toml")).
			Run()
		if err != nil {
			return err
		}
	} else {
		for idx, ip := range ips {
			who := fmt.Sprintf("%s@%s", username, ip)
			err := sh.Command("ssh", who, fmt.Sprintf(
				"sed -i '2c \"node1 %s:50011\"' ~/node%d/hosts.toml "+
					"&& sed -i '3c \"node2 %s:50012\"' ~/node%d/hosts.toml "+
					"&& sed -i '4c \"node3 %s:50013\"' ~/node%d/hosts.toml "+
					"&& sed -i '5c \"node4 %s:50014\"' ~/node%d/hosts.toml ",
				ips[0], idx+1,
				ips[1], idx+1,
				ips[2], idx+1,
				ips[3], idx+1)).
				Run()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func uploadFile(ips []string, username string, fpath string, fname string, cpath string, lpath string, lname string, ip_map map[string]int) error {
	for idx, ip := range ips {
		fmt.Printf("%s====> Uploading installation package and license to node%d%s\n", BLUE, idx+1, NC)
		v, _ := ip_map[ip]
		if v == idx+1 {
			ip_map[ip] = idx + 1
			who := fmt.Sprintf("%s@%s", username, ip)
			err := sh.
				Command("scp", lpath, fmt.Sprintf("%s:~/", who)).
				Command("scp", fpath, fmt.Sprintf("%s:~/", who)).
				Run()
			if err != nil {
				return err
			}
		} else {
			fmt.Printf("Node%d is on the same server as node%d, so the file has been uploaded.\n", idx+1, v)
		}
	}

	for idx, ip := range ips {
		fmt.Printf("%s====> Operating at node%d%s\n", BLUE, idx+1, NC)
		who := fmt.Sprintf("%s@%s", username, ip)
		err := sh.
			Command("ssh", who, fmt.Sprintf("mkdir -p ~/node%d && tar xzf %s -C ~/node%d --strip-components 1 && cp %s ~/node%d/LICENSE && cd ~/node%d && ./deploy-local.sh -d ~/node%d",
				idx+1,
				fname, idx+1,
				lname, idx+1,
				idx+1,
				idx+1)).
			Run()
		if err != nil {
			return err
		}
	}

	for idx, ip := range ips {
		fmt.Printf("%s====> Uploading certs to node%d%s\n", BLUE, idx+1, NC)
		who := fmt.Sprintf("%s@%s", username, ip)
		err := sh.
			Command("scp", "-r",
				fmt.Sprintf("%s/node%d", cpath, idx+1),
				fmt.Sprintf("%s:~/node%d/namespaces/global/config/certs", who, idx+1)).Run()
		if err != nil {
			return err
		}
	}

	return nil
}
