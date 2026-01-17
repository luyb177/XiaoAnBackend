package content

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

const (
	chunkSize     = 512 * 1024        // 512KB per chunk
	maxUploadSize = 100 * 1024 * 1024 // 100MB
)

// UploadContentStreamHandler 真正流式的 HTTP handler
func UploadContentStreamHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 给整个请求设置超时
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
		defer cancel()

		// 限制请求体大小
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

		// 获取 Content-Type 的 boundary
		contentType := r.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "multipart/") {
			httpx.OkJsonCtx(ctx, w, &types.Response{
				Code:    400,
				Message: "Content-Type 必须为 multipart/form-data",
			})
			return
		}
		boundaryIdx := strings.Index(contentType, "boundary=")
		if boundaryIdx == -1 {
			httpx.OkJsonCtx(ctx, w, &types.Response{
				Code:    400,
				Message: "multipart/form-data 缺少 boundary",
			})
			return
		}
		boundary := contentType[boundaryIdx+9:]

		// 建立 gRPC 流
		stream, err := svcCtx.ContentRpc.UploadContentStream(ctx)
		if err != nil {
			logx.Errorf("创建 gRPC 流失败: %v", err)
			httpx.OkJsonCtx(ctx, w, &types.Response{
				Code:    500,
				Message: "服务暂时不可用",
			})
			return
		}

		// 使用 multipart.NewReader 边读取边发送
		mr := multipart.NewReader(r.Body, boundary)
		var filename string
		isFirst := true
		totalSize := int64(0)

		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				logx.Errorf("读取表单 part 失败: %v", err)
				httpx.OkJsonCtx(ctx, w, &types.Response{
					Code:    500,
					Message: "解析文件失败",
				})
				return
			}
			if part.FileName() == "" {
				continue // 忽略非文件字段
			}
			filename = part.FileName()

			buf := make([]byte, chunkSize)
			for {
				n, err := part.Read(buf)
				if n > 0 {
					totalSize += int64(n)
					chunk := &content.UploadChunk{
						Data:   buf[:n],
						IsLast: false,
					}
					if isFirst {
						chunk.Filename = filename
						isFirst = false
					}
					if sendErr := stream.Send(chunk); sendErr != nil {
						logx.Errorf("发送 gRPC 分片失败: %v", sendErr)
						httpx.OkJsonCtx(ctx, w, &types.Response{
							Code:    500,
							Message: "上传失败",
						})
						return
					}
				}
				if err == io.EOF {
					break
				}
				if err != nil {
					logx.Errorf("读取文件分片失败: %v", err)
					httpx.OkJsonCtx(ctx, w, &types.Response{
						Code:    500,
						Message: "上传失败",
					})
					return
				}
			}
		}

		// 发送最后一个分片标记
		if err := stream.Send(&content.UploadChunk{
			Filename: filename,
			IsLast:   true,
		}); err != nil {
			logx.Errorf("发送结束标记失败: %v", err)
			httpx.OkJsonCtx(ctx, w, &types.Response{
				Code:    500,
				Message: "上传失败",
			})
			return
		}

		// 接收 gRPC 响应
		resp, err := stream.CloseAndRecv()
		if err != nil {
			logx.Errorf("接收 gRPC 响应失败: %v", err)
			httpx.OkJsonCtx(ctx, w, &types.Response{
				Code:    500,
				Message: "上传失败",
			})
			return
		}

		// 转换
		uploadResp := &content.UploadResponse{}
		if err := resp.Data.UnmarshalTo(uploadResp); err != nil {
			logx.Errorf("解析 gRPC 响应失败: %v", err)
			httpx.OkJsonCtx(ctx, w, &types.Response{
				Code:    500,
				Message: "响应解析失败",
			})
			return
		}

		logx.Infof("文件上传成功: filename=%s, size=%d, url=%s", filename, totalSize, uploadResp.Url)

		httpx.OkJsonCtx(ctx, w, &types.Response{
			Code:    200,
			Message: "上传成功",
			Data: &types.UploadContentResponse{
				Url: uploadResp.Url,
			},
		})
	}
}
