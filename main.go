package main

/*
#cgo CFLAGS: -I ./ -I./lib
#cgo CXXFLAGS: -I./
#cgo LDFLAGS:  -L./lib -lWeWorkFinanceSdk_C -ldl

#include "./lib/WeWorkFinanceSdk_C.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"github.com/gin-gonic/gin"
	val "github.com/go-playground/validator/v10"
	_ "github.com/pkg/errors"
	"golang.org/x/net/context"
	"msg/common/id_generator"
	"msg/common/log"
	"msg/common/storage"
	"msg/common/validator"
	"msg/conf"
	"msg/constants"
	"msg/controller"
	"msg/models"
	"msg/services"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// 调用微信的动态链接库，出现无法链接的错误需要手动添加库路径
// export LD_LIBRARY_PATH=${LD_LIBRARY_PATH}:<path-to-so>

func validateConfig(c interface{}) {
	if err := validator.NewCustomValidator().ValidateStruct(c); err != nil {
		panic(err.(val.ValidationErrors))
	}
}

func init() {
	err := conf.SetupSetting()
	if err != nil {
		panic(err)
	}
	log.SetupLogger(conf.Settings.App.Env)
	validateConfig(conf.Settings)
	models.DB = models.InitDB(conf.Settings.DB)
	id_generator.SetupIDGenerator()
	storage.Setup(conf.Settings.Storage)
}

func main() {
	arch := services.NewMsgArch()
	arch.Init()
	err := arch.Sync(conf.Settings.WeWork.ExtCorpID)
	if err != nil {
		panic(err)
	}

	msgArch := controller.NewMsgArch()
	r := gin.New()

	apiV1 := r.Group("/api/v1")
	apiV1.POST(constants.MsgArchSrvPathSync, msgArch.Sync)
	apiV1.GET(constants.MsgArchSrvPathSessions, msgArch.QuerySessions)
	apiV1.GET(constants.MsgArchSrvPathMsgs, msgArch.QueryChatMsgs)
	apiV1.GET(constants.MsgArchSrvSearchMsgs, msgArch.SearchMsgs)

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", conf.Settings.Server.MsgArchHttpPort),
		Handler:        r,
		ReadTimeout:    conf.Settings.Server.ReadTimeout,
		WriteTimeout:   conf.Settings.Server.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Sugar.Fatalf("s.ListenAndServe err: %v", err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Sugar.Fatalf("server forced to shutdown: %v", err)
	}

	log.Sugar.Info("Server exited")
}
