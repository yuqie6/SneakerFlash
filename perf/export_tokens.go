package main

import (
	"context"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type baseResp[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

type loginData struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type claimPayload struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
}

type tokenResult struct {
	idx      int
	token    string
	userID   uint
	username string
	err      error
}

func main() {
	baseURL := flag.String("base-url", "http://localhost:8000/api/v1", "API 基础地址，示例：http://localhost:8000/api/v1")
	prefix := flag.String("prefix", "perf_u", "用户名前缀，实际用户名为 <prefix>_idx")
	password := flag.String("password", "PerfTest#123", "注册/登录密码")
	count := flag.Int("count", 1000, "生成的用户数量")
	out := flag.String("out", "token.csv", "输出 CSV 路径")
	workers := flag.Int("workers", 20, "并发 worker 数量")
	flag.Parse()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	jobs := make(chan int, *workers)
	results := make(chan tokenResult, *count)

	var wg sync.WaitGroup
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				username := fmt.Sprintf("%s_%d", *prefix, idx)
				if err := register(client, *baseURL, username, *password); err != nil {
					results <- tokenResult{idx: idx, err: fmt.Errorf("register %s: %w", username, err)}
					continue
				}
				token, err := login(client, *baseURL, username, *password)
				if err != nil {
					results <- tokenResult{idx: idx, err: fmt.Errorf("login %s: %w", username, err)}
					continue
				}
				claim := decodeClaim(token)
				results <- tokenResult{
					idx:      idx,
					token:    token,
					userID:   claim.UserID,
					username: claim.Username,
				}
			}
		}()
	}

	go func() {
		for i := 0; i < *count; i++ {
			jobs <- i
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	records := make([]tokenResult, *count)
	for res := range results {
		if res.err != nil {
			log.Fatalf("生成 token 失败: %v", res.err)
		}
		records[res.idx] = res
	}

	if err := writeCSV(*out, records); err != nil {
		log.Fatalf("写入 csv 失败: %v", err)
	}
	log.Printf("完成，输出 %d 条记录到 %s\n", len(records), *out)
}

func register(client *http.Client, baseURL, username, password string) error {
	payload := map[string]string{
		"user_name":     username,
		"user_password": password,
	}
	return postJSON(client, fmt.Sprintf("%s/register", baseURL), payload, func(code int, body []byte) error {
		var resp baseResp[struct{}]
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("unmarshal: %w", err)
		}
		if code == http.StatusOK && (resp.Code == 200 || resp.Code == 10001) {
			return nil
		}
		return fmt.Errorf("unexpected code=%d resp=%s", code, string(body))
	})
}

func login(client *http.Client, baseURL, username, password string) (string, error) {
	payload := map[string]string{
		"user_name":     username,
		"user_password": password,
	}
	var token string
	err := postJSON(client, fmt.Sprintf("%s/login", baseURL), payload, func(code int, body []byte) error {
		var resp baseResp[loginData]
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("unmarshal: %w", err)
		}
		if code == http.StatusOK && resp.Code == 200 && resp.Data.AccessToken != "" {
			token = resp.Data.AccessToken
			return nil
		}
		return fmt.Errorf("unexpected code=%d resp=%s", code, string(body))
	})
	return token, err
}

func postJSON(client *http.Client, url string, payload any, check func(code int, body []byte) error) error {
	data, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, strings.NewReader(string(data)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return check(resp.StatusCode, body)
}

func decodeClaim(token string) claimPayload {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return claimPayload{}
	}
	payload, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return claimPayload{}
	}
	var c claimPayload
	_ = json.Unmarshal(payload, &c)
	return c
}

func writeCSV(path string, rows []tokenResult) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{"token", "user_id", "username"}); err != nil {
		return err
	}
	for _, r := range rows {
		if err := w.Write([]string{r.token, strconv.FormatUint(uint64(r.userID), 10), r.username}); err != nil {
			return err
		}
	}
	return w.Error()
}
