package logic

import (
	"bytes"
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"google.golang.org/protobuf/types/known/anypb"
	"io"
	"time"

	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadContentStreamLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUploadContentStreamLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadContentStreamLogic {
	return &UploadContentStreamLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UploadContentStream 上传文件
func (l *UploadContentStreamLogic) UploadContentStream(stream v1.ContentService_UploadContentStreamServer) error {
	var (
		fileBuffer bytes.Buffer
		filename   string
	)

	// 接受流式数据
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return stream.SendAndClose(&v1.Response{
				Code:    500,
				Message: fmt.Sprintf("接收分片失败: %v", err),
			})
		}

		// 初次赋值
		if filename == "" {
			filename = chunk.Filename
		}

		_, err = fileBuffer.Write(chunk.Data)
		if err != nil {
			return stream.SendAndClose(&v1.Response{
				Code:    500,
				Message: fmt.Sprintf("写入分片失败: %v", err),
			})
		}
		if chunk.IsLast {
			break
		}
	}

	objectName := fmt.Sprintf("%s-%s", time.Now().Format("20060102T150405"), filename)
	reader := bytes.NewReader(fileBuffer.Bytes())
	size := int64(fileBuffer.Len())

	// 上传文件
	info, err := l.svcCtx.MinioClient.PutObject(
		l.ctx,
		l.svcCtx.Config.MinioConf.ContentBucket,
		objectName,
		reader,
		size,
		minio.PutObjectOptions{ContentType: "application/octet-stream"},
	)
	if err != nil {
		return stream.SendAndClose(&v1.Response{
			Code:    500,
			Message: fmt.Sprintf("上传失败: %v", err),
		})
	}
	url := fmt.Sprintf("%s/%s/%s", l.svcCtx.Config.MinioConf.EndPoint, l.svcCtx.Config.MinioConf.ContentBucket, objectName)
	l.Infof("上传成功: %+v", info)

	// todo 异步保存文件信息

	// 构造返回数据
	res := &v1.UploadResponse{Url: url}

	resAny, err := anypb.New(res)
	if err != nil {
		return stream.SendAndClose(&v1.Response{
			Code:    500,
			Message: fmt.Sprintf("构造返回数据失败: %v", err),
		})
	}

	return stream.SendAndClose(&v1.Response{
		Code:    200,
		Message: "上传成功",
		Data:    resAny,
	})
}
