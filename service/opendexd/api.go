package opendexd

import (
	"context"
	"fmt"
	"github.com/opendexnetwork/opendex-docker-api/config"
	pb "github.com/opendexnetwork/opendex-docker-api/service/opendexd/opendexrpc"
	"github.com/opendexnetwork/opendex-docker-api/utils"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func (t *Service) ConfigureRouter(r *gin.RouterGroup) {
	r.GET("/v1/opendexd/getinfo", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), config.DefaultApiTimeout)
		defer cancel()
		resp, err := t.GetInfo(ctx)
		utils.HandleProtobufResponse(c, resp, err)
	})

	r.GET("/v1/opendexd/getbalance", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), config.DefaultApiTimeout)
		defer cancel()
		resp, err := t.GetBalance(ctx, "")
		utils.HandleProtobufResponse(c, resp, err)
	})

	r.GET("/v1/opendexd/getbalance/:currency", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), config.DefaultApiTimeout)
		defer cancel()
		resp, err := t.GetBalance(ctx, c.Param("currency"))
		utils.HandleProtobufResponse(c, resp, err)
	})

	r.GET("/v1/opendexd/tradehistory", func(c *gin.Context) {
		limitStr := c.DefaultQuery("limit", "0")
		limit, err := strconv.ParseUint(limitStr, 10, 32)
		if err != nil {
			msg := fmt.Sprintf("invalid limit: %s", err.Error())
			utils.JsonError(c, msg, http.StatusBadRequest)
			return
		}
		if limit < 0 {
			msg := fmt.Sprintf("invalid limit: %d", limit)
			utils.JsonError(c, msg, http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), config.DefaultApiTimeout)
		defer cancel()
		resp, err := t.GetTradeHistory(ctx, uint32(limit))
		utils.HandleProtobufResponse(c, resp, err)
	})

	r.GET("/v1/opendexd/tradinglimits", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), config.DefaultApiTimeout)
		defer cancel()
		resp, err := t.GetTradingLimits(ctx, "")
		utils.HandleProtobufResponse(c, resp, err)
	})

	r.GET("/v1/opendexd/tradinglimits/:currency", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), config.DefaultApiTimeout)
		defer cancel()
		resp, err := t.GetTradingLimits(ctx, c.Param("currency"))
		utils.HandleProtobufResponse(c, resp, err)
	})

	r.POST("/v1/opendexd/create", func(c *gin.Context) {
		var params CreateParams
		err := c.BindJSON(&params)
		if err != nil {
			utils.JsonError(c, err.Error(), http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), config.DefaultApiTimeout)
		defer cancel()
		resp, err := t.CreateNode(ctx, params.Password)
		utils.HandleProtobufResponse(c, resp, err)
	})

	r.POST("/v1/opendexd/restore", func(c *gin.Context) {
		var params RestoreParams
		err := c.BindJSON(&params)
		if err != nil {
			utils.JsonError(c, err.Error(), http.StatusBadRequest)
			return
		}

		lndBTC, err := ioutil.ReadFile(filepath.Join(params.BackupDir, "lnd-BTC"))
		lndLTC, err := ioutil.ReadFile(filepath.Join(params.BackupDir, "lnd-LTC"))
		opendexd, err := ioutil.ReadFile(filepath.Join(params.BackupDir, "opendexd"))

		ctx, cancel := context.WithTimeout(context.Background(), config.DefaultApiTimeout)
		defer cancel()
		resp, err := t.RestoreNode(
			ctx,
			params.Password,
			strings.Split(params.SeedMnemonic, " "),
			map[string][]byte{
				"BTC": lndBTC,
				"LTC": lndLTC,
			},
			opendexd,
		)
		utils.HandleProtobufResponse(c, resp, err)
	})

	r.POST("/v1/opendexd/unlock", func(c *gin.Context) {
		var params UnlockParams
		err := c.BindJSON(&params)
		if err != nil {
			utils.JsonError(c, err.Error(), http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), config.DefaultApiTimeout)
		defer cancel()
		resp, err := t.UnlockNode(ctx, params.Password)
		utils.HandleProtobufResponse(c, resp, err)
	})

	r.POST("/v1/opendexd/changepass", func(c *gin.Context) {
		var params ChangepasswordParams
		err := c.BindJSON(&params)
		if err != nil {
			utils.JsonError(c, err.Error(), http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), config.DefaultApiTimeout)
		defer cancel()
		resp, err := t.ChangePassword(ctx, params.NewPassword, params.OldPassword)
		// ignore file removal error here
		_ = os.Remove("/root/network/.password-unset")
		utils.HandleProtobufResponse(c, resp, err)
	})

	r.GET("/v1/opendexd/getmnemonic", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), config.DefaultApiTimeout)
		defer cancel()
		resp, err := t.GetMnemonic(ctx)
		utils.HandleProtobufResponse(c, resp, err)
	})

	r.GET("/v1/opendexd/listpairs", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), config.DefaultApiTimeout)
		defer cancel()
		resp, err := t.ListPairs(ctx)
		utils.HandleProtobufResponse(c, resp, err)
	})

	r.GET("/v1/opendexd/listorders", func(c *gin.Context) {
		var params ListOrdersParams
		err := c.BindQuery(&params)
		if err != nil {
			utils.JsonError(c, err.Error(), http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), config.DefaultApiTimeout)
		defer cancel()
		resp, err := t.ListOrders(ctx, params.PairId, pb.ListOrdersRequest_Owner(params.Owner), params.Limit, params.IncludeAliases)
		utils.HandleProtobufResponse(c, resp, err)
	})

	r.GET("/v1/opendexd/orderbook", func(c *gin.Context) {
		var params OrderBookParams
		err := c.BindQuery(&params)
		if err != nil {
			utils.JsonError(c, err.Error(), http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), config.DefaultApiTimeout)
		defer cancel()
		resp, err := t.OrderBook(ctx, params.PairId, params.Precision, params.Limit)
		utils.HandleProtobufResponse(c, resp, err)
	})

	r.POST("/v1/opendexd/placeorder", func(c *gin.Context) {
		var params PlaceOrderParams
		err := c.BindJSON(&params)
		if err != nil {
			utils.JsonError(c, err.Error(), http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), config.DefaultApiTimeout)
		defer cancel()
		resp, err := t.PlaceOrder(ctx, params.PairId, params.Side, params.Price, params.Quantity, params.OrderId, params.ReplaceOrderId, params.ImmediateOrCancel)
		utils.HandleProtobufResponse(c, resp, err)
	})

	r.POST("/v1/opendexd/removeorder", func(c *gin.Context) {
		var params RemoveOrderParams
		err := c.BindJSON(&params)
		if err != nil {
			utils.JsonError(c, err.Error(), http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), config.DefaultApiTimeout)
		defer cancel()
		resp, err := t.RemoveOrder(ctx, params.OrderId, params.Quantity)
		utils.HandleProtobufResponse(c, resp, err)
	})
}

type CreateParams struct {
	Password string `json:"password"`
}

type RestoreParams struct {
	Password     string `json:"password"`
	SeedMnemonic string `json:"seedMnemonic"`
	BackupDir    string `json:"backupDir"`
}

type UnlockParams struct {
	Password string `json:"password"`
}

type ChangepasswordParams struct {
	NewPassword string `json: "newPassword"`
	OldPassword string `json: "oldPassword"`
}

type ListOrdersParams struct {
	PairId         string `form:"pairId" json:"pairId"`
	Owner          uint32 `form:"owner" json:"owner"`
	Limit          uint32 `form:"limit" json:"limit"`
	IncludeAliases bool   `form:"includeAliases" json:"includeAliases"`
}

type OrderBookParams struct {
	PairId    string `form:"pairId" json:"pairId"`
	Precision int32  `form:"precision" json:"precision"`
	Limit     uint32 `form:"limit" json:"limit"`
}

type PlaceOrderParams struct {
	Price             float64      `json: "price"`
	Quantity          uint64       `json: "quantity"`
	PairId            string       `json: "pairId"`
	OrderId           string       `json: "orderId"`
	Side              pb.OrderSide `json: "side"`
	ReplaceOrderId    string       `json: "replaceOrderId"`
	ImmediateOrCancel bool         `json: "immediateOrCancel"`
}

type RemoveOrderParams struct {
	OrderId  string `json: "orderId"`
	Quantity uint64 `json: "quantity"`
}
