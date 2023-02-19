package packets

import (
	"encoding/json"

	"github.com/brocaar/lorawan"
)

/*
func CreatePullDataPacket(GatewayMACAddr lorawan.EUI64) []byte {

	packetBytes := GetHeader(TypePullData, GatewayMACAddr, 0)

	return packetBytes
} */

type PullDPacket struct {
	Header  []byte
	Payload PullDataPayload
}

type PullDataPayload struct {
	Stat *Stat `json:"stat,omitempty"`
}

func (p *PullDPacket) MarshalBinary() ([]byte, error) {

	JSONPayload, err := json.Marshal(p.Payload)

	if err != nil {
		return nil, err
	}

	out := append(p.Header, JSONPayload...)

	return out, nil
}

/* func CreatePullDataPacket(GatewayMACAddr lorawan.EUI64, stat Stat) ([]byte, error) {

	header := GetHeader(TypePullData, GatewayMACAddr, 0)

	payload := PullDataPayload{
		Stat: &stat,
	}

	pkt := PullDPacket{
		header,
		payload,
	}

	return pkt.MarshalBinary()

} */

func CreatePullDataPacket(GatewayMACAddr lorawan.EUI64) []byte {
	header := GetHeader(TypePullData, GatewayMACAddr, 0)

	return header
}
