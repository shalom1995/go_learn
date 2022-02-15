package service

import (
	"go_learn/pipeline模式/pipeline在怪物工程师中的应用/util"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"time"
)

type SignHandler struct {
	PubKeyPath string
	TxIdURL    string
	SignURL    string
	NFTType    string
	Head       map[string]string

	SuitAddress string
	SuitTokenID string
}

type pipeParam struct {
	EncodeHash string
	TxID       string
	Sign       string

	Err error
}

type resHash struct {
	Code    string `json:"code"`
	ReqHash string `json:"reqHash"`
}

type reqSig struct {
	Code string `json:"code"`
	Sig  string `json:"sig"`
}

type Trigger func([]string) <-chan pipeParam
type ProxyFunc func(in <-chan pipeParam) <-chan pipeParam

func (s *SignHandler) trigger(encodeHex []string) <-chan pipeParam {
	out := make(chan pipeParam)
	go func() {
		for _, e := range encodeHex {
			out <- pipeParam{
				EncodeHash: e,
			}
		}
		close(out)
	}()

	return out
}

func (s *SignHandler) requestSignID(in <-chan pipeParam) <-chan pipeParam {
	out := make(chan pipeParam)
	go func() {
		for param := range in {
			timStruct := fmt.Sprintf(
				`{"contract":"%s", "type":"%s","tokenId":%s, "amount":0}`,
				s.SuitAddress, s.NFTType, s.SuitTokenID)
			signParam, err := GenSignRequest(param.EncodeHash, timStruct, s.PubKeyPath)
			if err != nil && param.Err == nil {
				param.Err = errors.Wrap(err, "GenSignRequest error")
			}

			res, err := util.PostUrlRetry(s.TxIdURL, nil, signParam, s.Head)
			if err != nil && param.Err == nil {
				param.Err = errors.Wrap(err, "PostUrlRetry error")
			}
			r := &resHash{}
			err = json.Unmarshal(res, r)
			if err != nil && param.Err == nil {
				param.Err = errors.Wrap(err, "requestSignID json.Unmarshal error")
			}
			param.TxID = r.ReqHash
			out <- param
		}
		close(out)
	}()
	return out
}

func (s *SignHandler) requestSign(in <-chan pipeParam) <-chan pipeParam {
	out := make(chan pipeParam)

	go func() {
		for param := range in {
			ticker := time.NewTicker(time.Second * 3)
			isContinue := true
			for isContinue {
				select {
				case <-ticker.C:
					d, err := util.GetURLRetry(s.SignURL+param.TxID, nil)
					if err != nil && param.Err == nil {
						param.Err = errors.Wrap(err, "GetURLRetry error")
					}
					rs := &reqSig{}
					err = json.Unmarshal(d, rs)
					if err != nil && param.Err == nil {
						param.Err = errors.Wrap(err, "requestSign json.Unmarshal error")
					}

					if rs.Sig != "pending" {
						out <- pipeParam{
							Sign: rs.Sig,
						}
						isContinue = false
					}
				}
			}
		}
		close(out)
	}()

	return out
}

func (s *SignHandler) pipeline(encodeHex []string, trigger Trigger, proxies ...ProxyFunc) <-chan pipeParam {
	ch := trigger(encodeHex)
	for i, _ := range proxies {
		ch = proxies[i](ch)
	}
	return ch
}

const keyWid = 8

type PRequestSignature struct {
	Timestamp  string `json:"timeStamp"`
	EncodeData string `json:"encodeData"`
	AssetsData string `json:"assetsData"`
	Key        string `json:"key"`
}

/**
 * @parameter: paramHash 合约参数encode后的hash；timStruct Tim要求的数据结构；path 公钥文件路径
 * @return: 构造好的参数 错误信息
 * @Description: 生成与签名机交互前的参数
 * @author: shalom
 * @date: 2022/1/27 3:28 下午
 * @version: V1.0
 */
func GenSignRequest(paramHash, timStruct, path string) (PRequestSignature, error) {
	r := PRequestSignature{}
	key := util.GenCode(keyWid)

	privateKey := util.GenBytesPrivateKey(key)
	encodeData, err := util.AESEncrypt(privateKey, []byte(paramHash))
	if err != nil {
		return r, errors.Wrap(err, "AESEncrypt paramHash error")
	}
	r.EncodeData = encodeData

	assetData, err := util.AESEncrypt(privateKey, []byte(timStruct))
	if err != nil {
		return r, errors.Wrap(err, "AESEncrypt signData error")
	}
	r.AssetsData = assetData

	r.Timestamp = fmt.Sprintf("%v", time.Now().Unix())

	pubKey, err := util.ParseRSAPubKey(path)
	if err != nil {
		return r, errors.Wrap(err, "ParseRSAPubKey error")
	}

	k, err := util.RSAPKCS1V15Encrypt(pubKey, []byte(key))
	if err != nil {
		return r, errors.Wrap(err, "RSAPKCS1V15Encrypt error")
	}
	r.Key = k
	return r, nil
}