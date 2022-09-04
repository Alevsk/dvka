package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/go-ping/ping"
	"github.com/gorilla/mux"
	"github.com/lixiangzhong/dnsutil"
	"github.com/miekg/dns"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

func app(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		http.ServeFile(w, r, "static/index.html")
	default:
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Sorry, only GET method is supported.")
	}
}

func admin(w http.ResponseWriter, r *http.Request) {
	rawIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Sorry, something went wrong.")
		return
	}

	proxyAddresses := r.Header.Values("X-Forwarded-For")

	if len(proxyAddresses) > 0 {
		rawIP = proxyAddresses[0]
	}

	if IsIPv6(rawIP) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "IPV6 verification not implemented yet, will be available in the next release, come back later")
		return
	}

	clientIP := net.ParseIP(rawIP)

	podIP := GetPodIP()
	podAddress := net.ParseIP(podIP)
	mask := podAddress.DefaultMask()
	network := podAddress.Mask(mask)

	_, podIPNet, _ := net.ParseCIDR(network.String() + "/8")
	var response string

	if podIPNet.Contains(clientIP) {
		response += `
 __ __ __ __ __   __  __         ___ __ __  
 /\ /  /  |_ (_ (_   / _ |__) /\ |\ | | |_ |  \ 
/--\\__\__|____)__)  \__)| \ /--\| \| | |__|__/ 

`
		response += fmt.Sprintf("flag: %v\n", GetFlag())

	} else {
		response += `
 ▄████▄   ▒█████   ███▄    █  ███▄    █ ▓█████  ▄████▄  ▄▄▄█████▓ ██▓ ▒█████   ███▄    █    ▄▄▄█████▓▓█████  ██▀███   ███▄ ▄███▓ ██▓ ███▄    █  ▄▄▄     ▄▄▄█████▓▓█████ ▓█████▄ 
▒██▀ ▀█  ▒██▒  ██▒ ██ ▀█   █  ██ ▀█   █ ▓█   ▀ ▒██▀ ▀█  ▓  ██▒ ▓▒▓██▒▒██▒  ██▒ ██ ▀█   █    ▓  ██▒ ▓▒▓█   ▀ ▓██ ▒ ██▒▓██▒▀█▀ ██▒▓██▒ ██ ▀█   █ ▒████▄   ▓  ██▒ ▓▒▓█   ▀ ▒██▀ ██▌
▒▓█    ▄ ▒██░  ██▒▓██  ▀█ ██▒▓██  ▀█ ██▒▒███   ▒▓█    ▄ ▒ ▓██░ ▒░▒██▒▒██░  ██▒▓██  ▀█ ██▒   ▒ ▓██░ ▒░▒███   ▓██ ░▄█ ▒▓██    ▓██░▒██▒▓██  ▀█ ██▒▒██  ▀█▄ ▒ ▓██░ ▒░▒███   ░██   █▌
▒▓▓▄ ▄██▒▒██   ██░▓██▒  ▐▌██▒▓██▒  ▐▌██▒▒▓█  ▄ ▒▓▓▄ ▄██▒░ ▓██▓ ░ ░██░▒██   ██░▓██▒  ▐▌██▒   ░ ▓██▓ ░ ▒▓█  ▄ ▒██▀▀█▄  ▒██    ▒██ ░██░▓██▒  ▐▌██▒░██▄▄▄▄██░ ▓██▓ ░ ▒▓█  ▄ ░▓█▄   ▌
▒ ▓███▀ ░░ ████▓▒░▒██░   ▓██░▒██░   ▓██░░▒████▒▒ ▓███▀ ░  ▒██▒ ░ ░██░░ ████▓▒░▒██░   ▓██░     ▒██▒ ░ ░▒████▒░██▓ ▒██▒▒██▒   ░██▒░██░▒██░   ▓██░ ▓█   ▓██▒ ▒██▒ ░ ░▒████▒░▒████▓ 
░ ░▒ ▒  ░░ ▒░▒░▒░ ░ ▒░   ▒ ▒ ░ ▒░   ▒ ▒ ░░ ▒░ ░░ ░▒ ▒  ░  ▒ ░░   ░▓  ░ ▒░▒░▒░ ░ ▒░   ▒ ▒      ▒ ░░   ░░ ▒░ ░░ ▒▓ ░▒▓░░ ▒░   ░  ░░▓  ░ ▒░   ▒ ▒  ▒▒   ▓▒█░ ▒ ░░   ░░ ▒░ ░ ▒▒▓  ▒ 
  ░  ▒     ░ ▒ ▒░ ░ ░░   ░ ▒░░ ░░   ░ ▒░ ░ ░  ░  ░  ▒       ░     ▒ ░  ░ ▒ ▒░ ░ ░░   ░ ▒░       ░     ░ ░  ░  ░▒ ░ ▒░░  ░      ░ ▒ ░░ ░░   ░ ▒░  ▒   ▒▒ ░   ░     ░ ░  ░ ░ ▒  ▒ 
░        ░ ░ ░ ▒     ░   ░ ░    ░   ░ ░    ░   ░          ░       ▒ ░░ ░ ░ ▒     ░   ░ ░      ░         ░     ░░   ░ ░      ░    ▒ ░   ░   ░ ░   ░   ▒    ░         ░    ░ ░  ░ 
░ ░          ░ ░           ░          ░    ░  ░░ ░                ░      ░ ░           ░                ░  ░   ░            ░    ░           ░       ░  ░           ░  ░   ░    
░                                              ░                                                                                                                         ░      
`

		response += fmt.Sprintf("[+] Advanced Web Application Firewall (WAF) 1.0\n")
		response += fmt.Sprintf("[+] Requested resource: %v\n", podIP)
		response += fmt.Sprintf("===============================================\n")
		response += fmt.Sprintf("[+] Attack stopped!\n")
		response += fmt.Sprintf("[+] Malicios IP detected: %v\n", rawIP)

		//response += fmt.Sprintf("proxies: %v\n", proxyAddresses)
		//response += fmt.Sprintf("client IP: %v\n", rawIP)
		//response += fmt.Sprintf("Client IP: %v\n", clientIP.String())
		//response += fmt.Sprintf("pod IP: %v\n", podIP)
		//response += fmt.Sprintf("mask: %v\n", mask)
		//response += fmt.Sprintf("network: %v\n", network)

		response += fmt.Sprintf("[+] Reason: only internal network connections are allowed\n")
	}

	fmt.Fprintf(w, response)
}

func apiV1(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		switch r.URL.Path {
		case "/api/v1/run":
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Sorry, something went wrong.")
				return
			}
			r.Body.Close()
			var params RunCommandRequest
			err = json.Unmarshal(body, &params)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Sorry, something went wrong.")
				return
			}

			var command string
			var hostname string
			var bodyResponse []byte

			if params.Command == "" {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Sorry, something went wrong.")
				return
			}
			if params.Hostname == "" {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Sorry, something went wrong.")
				return
			}

			command = params.Command
			hostname = params.Hostname

			switch command {
			case "ping":
				pinger, err := ping.NewPinger(hostname)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "Sorry, something went wrong.")
					return
				}
				pinger.OnRecv = func(pkt *ping.Packet) {
					bodyResponse = append(bodyResponse, []byte(fmt.Sprintf("%d bytes from %s: icmp_seq=%d time=%v\n",
						pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt))...)
				}
				pinger.OnDuplicateRecv = func(pkt *ping.Packet) {
					bodyResponse = append(bodyResponse, []byte(fmt.Sprintf("%d bytes from %s: icmp_seq=%d time=%v ttl=%v (DUP!)\n",
						pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt, pkt.Ttl))...)
				}

				pinger.OnFinish = func(stats *ping.Statistics) {
					bodyResponse = append(bodyResponse, []byte(fmt.Sprintf("\n--- %s ping statistics ---\n", stats.Addr))...)
					bodyResponse = append(bodyResponse, []byte(fmt.Sprintf("%d packets transmitted, %d packets received, %v%% packet loss\n",
						stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss))...)
					bodyResponse = append(bodyResponse, []byte(fmt.Sprintf("round-trip min/avg/max/stddev = %v/%v/%v/%v\n",
						stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt))...)
				}
				bodyResponse = append(bodyResponse, []byte(fmt.Sprintf("PING %s (%s):\n", pinger.Addr(), pinger.IPAddr()))...)
				pinger.Count = 3
				if err = pinger.Run(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "Sorry, something went wrong.")
					return
				}
				fmt.Fprintf(w, string(bodyResponse))

			case "curl":
				client := &http.Client{
					Transport: &http.Transport{
						TLSClientConfig: &tls.Config{
							InsecureSkipVerify: true,
						},
					},
				}

				curlReq, err := http.NewRequest(r.Method, hostname, nil)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "Sorry, something went wrong.")
					return
				}

				curlReq.Header = r.Header
				proxies := curlReq.Header.Values("x-forwarded-for")
				if len(proxies) == 0 {
					curlReq.Header.Set("x-forwarded-for", r.RemoteAddr)
				}

				resp, err := client.Do(curlReq)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "Sorry, something went wrong.")
					return
				}
				bodyResponse, err = ioutil.ReadAll(resp.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "Sorry, something went wrong.")
					return
				}
				resp.Body.Close()
				fmt.Fprintf(w, string(bodyResponse))

			case "nslookup":
				ips, err := net.LookupIP(hostname)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "Sorry, something went wrong.")
					return
				}
				for _, ip := range ips {
					bodyResponse = append(bodyResponse, []byte(fmt.Sprintf("%s. IN A %s\n", hostname, ip.String()))...)
				}
				fmt.Fprintf(w, string(bodyResponse))

			case "dig":
				var dig dnsutil.Dig
				msg, err := dig.GetMsg(dns.TypeA, hostname)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "Sorry, something went wrong.")
					return
				}
				bodyResponse = append(bodyResponse, []byte(fmt.Sprintf("Question: %v. \n", msg.Question))...)
				bodyResponse = append(bodyResponse, []byte(fmt.Sprintf("Answer: %s. \n", msg.Answer))...)
				bodyResponse = append(bodyResponse, []byte(fmt.Sprintf("Ns: %s. \n", msg.Ns))...)
				bodyResponse = append(bodyResponse, []byte(fmt.Sprintf("Extra: %s. \n", msg.Extra))...)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "Sorry, something went wrong.")
					return
				}
				fmt.Fprintf(w, string(bodyResponse))

			default:
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Sorry, unsupported api.")
			}

		default:
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Sorry, unsupported api.")
		}
	default:
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Sorry, unsupported method .")
	}
}

func main() {
	r := mux.NewRouter()
	r.UseEncodedPath()

	if GetAdminPanel() == "on" {
		r.PathPrefix("/admin").HandlerFunc(admin)
	}

	r.PathPrefix("/assets").Handler(http.FileServer(http.Dir("./static/")))
	r.PathPrefix("/api/v1").HandlerFunc(apiV1)
	r.HandleFunc("/", app)

	if len(os.Args) > 2 {
		log.Fatal("Usage: ./lab2 addr:port")
	}

	addr := "0.0.0.0:8080"

	if len(os.Args) == 2 {
		addr = os.Args[1]
	}

	fmt.Printf("Starting server at %s\n", addr)

	srv := &http.Server{
		Handler: r,
		Addr:    addr,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 3600 * time.Second,
		ReadTimeout:  3600 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
