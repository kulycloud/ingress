package servers

import (
	"errors"
	"github.com/google/uuid"
	commonHttp "github.com/kulycloud/common/http"
	protoHttp "github.com/kulycloud/protocol/http"
	protoStorage "github.com/kulycloud/protocol/storage"
	"io"
	"net/http"
)

// wrapper for the header
type protoRequestHeader protoHttp.RequestHeader

func newProtoRequestHeader() *protoRequestHeader {
	return &protoRequestHeader{
		ServiceData: make(commonHttp.ServiceData),
	}
}

func (prh *protoRequestHeader) setHttpData(r *http.Request) {
	prh.HttpData = &protoHttp.RequestHeader_HttpData{
		Method: r.Method,
		Path:   r.URL.Path,
		Source: r.RemoteAddr,
	}
	headers := make(commonHttp.Headers)
	for key, values := range r.Header {
		headers.SetValues(key, values)
	}
	prh.HttpData.Headers = headers
}

func (prh *protoRequestHeader) setKulyData(s *protoStorage.GetRouteStartResponse) {
	prh.KulyData = &protoHttp.RequestHeader_KulyData{
		RouteUid:   s.Uid,
		StepUid:    0,
		Step:       s.Step,
		RequestUid: uuid.New().String(),
	}
}

func (prh *protoRequestHeader) toChunk() *protoHttp.Chunk {
	return &protoHttp.Chunk{
		Content: &protoHttp.Chunk_Header{
			Header: &protoHttp.Header{
				Content: &protoHttp.Header_RequestHeader{
					RequestHeader: (*protoHttp.RequestHeader)(prh),
				},
			},
		},
	}
}

func bodyToChunks(body io.ReadCloser) <-chan *protoHttp.Chunk {
	chunks := make(chan *protoHttp.Chunk, 1)
	go func() {
		defer body.Close()
		buffer := make([]byte, commonHttp.MaxChunkSize)
		for {
			count, err := body.Read(buffer)
			if count > 0 {
				chunks <- &protoHttp.Chunk{
					Content: &protoHttp.Chunk_BodyChunk{
						BodyChunk: buffer[:count],
					},
				}
			}
			if err != nil {
				if !errors.Is(err, io.EOF) {
					logger.Errorw("error during body parsing", "error", err)
				}
				break
			}
		}
		close(chunks)
	}()
	return chunks
}

func sendRequest(stream protoHttp.Http_ProcessRequestClient, r *http.Request, routeStart *protoStorage.GetRouteStartResponse) error {
	header := newProtoRequestHeader()
	header.setHttpData(r)
	header.setKulyData(routeStart)
	err := stream.Send(header.toChunk())
	if err != nil {
		return err
	}
	chunks := bodyToChunks(r.Body)
	for {
		chunk, ok := <-chunks
		if !ok {
			break
		}
		err := stream.Send(chunk)
		if err != nil {
			return err
		}
	}
	return stream.CloseSend()
}

func writeResponse(stream protoHttp.Http_ProcessRequestClient, w http.ResponseWriter) error {
	chunk, err := stream.Recv()
	if err != nil {
		return err
	}
	header := chunk.GetHeader().GetResponseHeader()
	err = writeResponseHeader(header, w)
	if err != nil {
		return err
	}
	for {
		chunk, err = stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		_, err = w.Write(chunk.GetBodyChunk())
		if err != nil {
			return err
		}
	}
}

func writeResponseHeader(header *protoHttp.ResponseHeader, w http.ResponseWriter) error {
	if header == nil {
		return commonHttp.ErrConversionError
	}
	for key, value := range header.Headers {
		w.Header().Set(key, value)
	}
	w.Header().Set("Request-Uid", header.RequestUid)
	w.WriteHeader(int(header.Status))
	return nil
}
