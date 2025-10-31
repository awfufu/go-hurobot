// qbot/qbot.go
package qbot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"go-hurobot/config"
)

func NewClient() *Client {
	client := &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	// 启动反向 HTTP 服务器
	go client.startHTTPServer()

	log.Printf("正向 HTTP 地址: %s", config.Cfg.NapcatHttpServer)
	log.Printf("反向 HTTP 监听: http://%s", config.Cfg.ReverseHttpServer)

	return client
}

func (c *Client) Close() {
	if c.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := c.server.Shutdown(ctx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
	}
}

// 启动反向 HTTP 服务器，接收 NapCat 推送的消息
func (c *Client) startHTTPServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", c.handleHTTPEvent)

	c.server = &http.Server{
		Addr:         config.Cfg.ReverseHttpServer,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := c.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}
}

// 处理 NapCat 推送的事件
func (c *Client) handleHTTPEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("读取请求体失败: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// 解析 JSON
	jsonMap := make(map[string]any)
	if err := json.Unmarshal(body, &jsonMap); err != nil {
		log.Printf("解析 JSON 失败: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 处理事件
	if postType, exists := jsonMap["post_type"]; exists {
		if str, ok := postType.(string); ok && str != "" {
			go c.handleEvents(&str, &body, &jsonMap)
		}
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// 发送 API 请求到 NapCat（正向 HTTP）
// 统一的 HTTP 请求方法
func (c *Client) sendRequest(req *cqRequest) (*http.Response, error) {
	jsonBytes, err := json.Marshal(req.Params)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest(http.MethodPost, config.Cfg.NapcatHttpServer+"/"+req.Action, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if config.Cfg.ApiKeys.Longport.AccessToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+config.Cfg.ApiKeys.Longport.AccessToken)
	}

	return c.httpClient.Do(httpReq)
}

func (c *Client) sendJson(req *cqRequest) error {
	resp, err := c.sendRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *Client) sendWithResponse(req *cqRequest) (*cqResponse, error) {
	resp, err := c.sendRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	var cqResp cqResponse
	if err := json.Unmarshal(body, &cqResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return &cqResp, nil
}

// 发送请求并返回 JSON 字符串（用于测试 API）
func (c *Client) sendWithJSONResponse(req *cqRequest) (string, error) {
	resp, err := c.sendRequest(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("API 响应: %s", string(body))
	return string(body), nil
}
