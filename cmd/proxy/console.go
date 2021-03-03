package main

import (
	"encoding/json"
	"fmt"
	"github.com/creack/pty"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	socketio "github.com/googollee/go-socket.io"
	"net/http"
	"os"
	"os/exec"
	"sync"
)

var (
	consoleMap = make(map[string]Console)
	mutex      = sync.Mutex{}
)

type Console struct {
	Id           string   `json:"id"`
	Network      string   `json:"network"`
	ConnectionId string   `json:"connectionId"`
	Pty          *os.File `json:"-"`
}

func init() {
	router.GET("/api/v1/consoles", listConsoles)
	router.GET("/api/v1/consoles/:id", getConsole)
}

func findById(id string) *Console {
	console, ok := consoleMap[id]
	if !ok {
		return nil
	}
	return &console
}

func findIdsByConnectionId(connectionId string) []string {
	consoles := make([]string, 0)
	for key, value := range consoleMap {
		if value.ConnectionId == connectionId {
			consoles = append(consoles, key)
		}
	}
	return consoles
}

func getConsole(c *gin.Context) {
	id := c.Param("id")
	console := findById(id)
	if console == nil {
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}
	c.JSON(http.StatusOK, console)
}

func listConsoles(c *gin.Context) {
	c.JSON(http.StatusOK, consoleMap)
}

type TerminalSize struct {
	Rows uint16 `json:"rows"`
	Cols uint16 `json:"cols"`
}

type StartRequest struct {
	Id   string       `json:"id"`
	Size TerminalSize `json:"size"`
}

type ResizeRequest struct {
	Id   string       `json:"id"`
	Size TerminalSize `json:"size"`
}

type StopRequest struct {
	Id string `json:"id"`
}

func removeConsoles(connectionId string) {
	consolesToRemove := findIdsByConnectionId(connectionId)
	if len(consolesToRemove) > 0 {
		for _, val := range consolesToRemove {
			delete(consoleMap, val)
		}
	}
}

var help = `\
Opendex-cli shortcut commands
  addcurrency <currency>                    add a currency
  <swap_client> [decimal_places]
  [token_address]
  addpair <pair_id|base_currency>           add a trading pair
  [quote_currency]
  ban <node_identifier>                     ban a remote node
  buy <quantity> <pair_id> <price>          place a buy order
  [order_id]
  closechannel <currency>                   close any payment channels with a
  [node_identifier ] [--force]              peer
  connect <node_uri>                        connect to a remote node
  create                                    create a new opendexd instance and set a
                                            password
  discovernodes <node_identifier>           discover nodes from a specific peer
  getbalance [currency]                     get total balance for a given
                                            currency
  getinfo                                   get general info from the local opendexd
                                            node
  getnodeinfo <node_identifier>             get general information about a
                                            known node
  listcurrencies                            list available currencies
  listorders [pair_id] [owner]              list orders from the order book
  [limit]
  listpairs                                 get order book's available pairs
  listpeers                                 list connected peers
  openchannel <currency> <amount>           open a payment channel with a peer
  [node_identifier] [push_amount]
  orderbook [pair_id] [precision]           display the order book, with orders
                                            aggregated per price point
  removecurrency <currency>                 remove a currency
  removeorder <order_id> [quantity]         remove an order
  removepair <pair_id>                      remove a trading pair
  restore [backup_directory]                restore an opendexd instance from seed
  sell <quantity> <pair_id> <price>         place a sell order
  [order_id]
  shutdown                                  gracefully shutdown local opendexd node
  streamorders [existing]                   stream order added, removed, and
                                            swapped events (DEMO)
  tradehistory [limit]                      list completed trades
  tradinglimits [currency]                  trading limits for a given currency
  unban <node_identifier>                   unban a previously banned remote
  [--reconnect]                             node
  unlock                                    unlock local opendexd node
  walletdeposit <currency>                  gets an address to deposit funds to
                                            opendexd
  walletwithdraw [amount] [currency]        withdraws on-chain funds from opendexd
  [destination] [fee]
  
General commands
  report                                    report issue
  logs                                      show service log
  start                                     start service
  stop                                      stop service
  restart                                   restart service
  up                                        bring up the environment
  help                                      show this help
  exit                                      exit opendexd-ctl shell

CLI commands
  bitcoin-cli                               bitcoind cli
  litecoin-cli                              litecoind cli
  lndbtc-lncli                              lnd cli
  lndltc-lncli                              lnd cli
  geth                                      geth cli
  opendex-cli                                     opendexd cli
  boltzcli                                  boltz cli

Boltzcli shortcut commands  
  boltzcli <chain> deposit 
  --inbound [inbound_balance]               deposit from boltz (btc/ltc)
  boltzcli <chain> withdraw 
  <amount> <address>                        withdraw from boltz channel
`

func writeInitScript(network string) {
	f, err := os.Create("init.bash")
	if err != nil {
		logger.Errorf("Failed to write init.bash: %s", err)
		return
	}
	defer f.Close()

	f.WriteString(`\
cat <<EOF
             .___                         __  .__   
           __| _/____ ___  ___      _____/  |_|  |  
          / __ |/ __ \\  \/  /    _/ ___\   __\  |  
         / /_/ \  ___/ >    <     \  \___|  | |  |__
         \____ |\___  >__/\_ \     \___  >__| |____/
              \/    \/      \/         \/           

--------------------------------------------------------------

EOF

export NETWORK=` + network + `
export PS1="$NETWORK > "
function help() {
	echo "` + help + `"
}
function start() {
	docker start ${NETWORK}_${1}_1 
}
function stop() {
	docker stop ${NETWORK}_${1}_1
}
function restart() {
	docker restart ${NETWORK}_${1}_1
}
function down() {
	echo "Not implemented yet!"
}
function logs() {
	docker logs --tail=100 ${NETWORK}_${1}_1
}
function report() {
	cat <<EOF
Please click on https://github.com/opendexnetwork/opendexd/issues/\
new?assignees=kilrau&labels=bug&template=bug-report.md&title=Short%2C+concise+\
description+of+the+bug, describe your issue, drag and drop the file "${NETWORK}\
.log" which is located in "{logs_dir}" into your browser window and submit \
your issue.
EOF
}
function opendex-cli() {
	docker exec -it ${NETWORK}_opendexd_1 opendex-cli $@
}
function lndbtc-lncli() {
	docker exec -it ${NETWORK}_lndbtc_1 lncli -n ${NETWORK} -c bitcoin $@
}
function lndltc-lncli() {
	docker exec -it ${NETWORK}_lndltc_1 lncli -n ${NETWORK} -c litecoin $@
}
function geth() {
	docker exec -it ${NETWORK}_geth_1 geth $@
}
function bitcoin-ctl() {	
	if [[ $NETWORK == "testnet" ]]; then
		docker exec -it ${NETWORK}_bitcoind_1 -testnet -user xu -password xu bitcoind $@
	else
		docker exec -it ${NETWORK}_bitcoind_1 -user xu -password xu bitcoind $@
	fi
}
function litecoin-ctl() {
	if [[ $NETWORK == "testnet" ]]; then
		docker exec -it ${NETWORK}_litecoind_1 -testnet -user xu -password xu litecoind $@
	else
		docker exec -it ${NETWORK}_litecoind_1 -user xu -password xu litecoind $@
	fi
}
function boltzcli() {
	docker exec -it ${NETWORK}_boltz_1 wrapper $@
}

alias getinfo='opendex-cli getinfo'
alias addcurrency='opendex-cli addcurrency'
alias addpair='opendex-cli addpair'
alias ban='opendex-cli ban'
alias buy='opendex-cli buy'
alias closechannel='opendex-cli closechannel'
alias connect='opendex-cli connect'
alias create='opendex-cli create'
alias discovernodes='opendex-cli discovernodes'
alias getbalance='opendex-cli getbalance'
alias getnodeinfo='opendex-cli getnodeinfo'
alias listcurrencies='opendex-cli listcurrencies'
alias listorders='opendex-cli listorders'
alias listpairs='opendex-cli listpairs'
alias listpeers='opendex-cli listpeers'
alias openchannel='opendex-cli openchannel'
alias orderbook='opendex-cli orderbook'
alias removeallorders='opendex-cli removeallorders'
alias removecurrency='opendex-cli removecurrency'
alias removeorder='opendex-cli removeorder'
alias removepair='opendex-cli removepair'
alias restore='opendex-cli restore'
alias sell='opendex-cli sell'
alias shutdown='opendex-cli shutdown'
alias streamorders='opendex-cli streamorders'
alias tradehistory='opendex-cli tradehistory'
alias tradinglimits='opendex-cli tradinglimits'
alias unban='opendex-cli unban'
alias unlock='opendex-cli unlock'
alias walletdeposit='opendex-cli walletdeposit'
alias walletwithdraw='opendex-cli walletwithdraw'
`)
}

func startShell(console *Console, size TerminalSize) error {
	writeInitScript(console.Network)
	c := exec.Command("/bin/bash", "--init-file", "init.bash")

	ptmx, err := pty.StartWithSize(c, &pty.Winsize{Cols: size.Cols, Rows: size.Rows, X: 0, Y: 0})
	if err != nil {
		return err
	}

	console.Pty = ptmx

	return nil
}

func initSioConsole() {
	sioServer.OnEvent("/", "create", func(s socketio.Conn, data string) {
		network := os.Getenv("NETWORK")
		id := fmt.Sprint(uuid.New())
		console := Console{
			Id:           id,
			Network:      network,
			ConnectionId: s.ID(),
		}
		consoleMap[id] = console
		e := fmt.Sprintf("console-%s", id)
		logger.Debugf("[console] Register event %s", e)
		sioServer.BroadcastToRoom("/", s.ID(), "created", console.Id)
	})

	sioServer.OnEvent("/", "start", func(s socketio.Conn, data string) {
		req := StartRequest{}
		err := json.Unmarshal([]byte(data), &req)
		if err != nil {
			s.Emit("start", fmt.Sprintf("invalid request: %s", err))
			return
		}

		logger.Debugf("[console] Start %s", req.Id)

		console := findById(req.Id)
		if console == nil {
			s.Emit("start", "console not found")
			return
		}
		err = startShell(console, req.Size)
		if err != nil {
			s.Emit("start", fmt.Sprintf("failed to start: %s", err))
			return
		}

		inputEvent := fmt.Sprintf("console.%s.input", console.Id)
		outputEvent := fmt.Sprintf("console.%s.output", console.Id)

		sioServer.OnEvent("/", inputEvent, func(s socketio.Conn, data string) {
			logger.Debugf("[console/%s] ---> %v", console.Id, data)

			pty_ := console.Pty

			_, err := pty_.WriteString(data)
			if err != nil {
				logger.Errorf("Failed to write to console %s: %s", console.Id, err)
			}
		})

		go func() {
			var buf = make([]byte, 65536)
			for {
				pty_ := console.Pty

				n, err := pty_.Read(buf)
				if err != nil {
					logger.Errorf("Failed to read from console %s: %s", req.Id, err)
					break
				}
				data := buf[:n]
				logger.Debugf("[console/%s] <--- %v", console.Id, data)

				sioServer.BroadcastToRoom("/", s.ID(), outputEvent, string(data))
			}
		}()
	})

	sioServer.OnEvent("/", "resize", func(s socketio.Conn, data string) {
		req := ResizeRequest{}
		err := json.Unmarshal([]byte(data), &req)
		if err != nil {
			s.Emit("resize", fmt.Sprintf("invalid request: %s", err))
			return
		}

		console := findById(req.Id)
		if console == nil {
			s.Emit("resize", "console not found")
			return
		}

		err = pty.Setsize(console.Pty, &pty.Winsize{Rows: req.Size.Rows, Cols: req.Size.Cols, X: 0, Y: 0})
		if err != nil {
			s.Emit("resize", fmt.Sprintf("failed to resize: %s", err))
		}
	})

	sioServer.OnEvent("/", "stop", func(s socketio.Conn, data string) {
		req := StopRequest{}
		err := json.Unmarshal([]byte(data), &req)
		if err != nil {
			s.Emit("stop", fmt.Sprintf("invalid request: %s", err))
			return
		}

		delete(consoleMap, req.Id)
	})
}
