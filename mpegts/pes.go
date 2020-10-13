package mpegts

// isPESPayload checks whether the payload is a PES one
func isPESPayload(i []byte) bool {
	// Packet is not big enough
	if len(i) < 3 {
		return false
	}

	// Check prefix
	return uint32(i[0])<<16|uint32(i[1])<<8|uint32(i[2]) == 1
}

func EncodeSPSPPS(pid uint16) ([]byte, error) {
	pkt := Packet{
		PID:     pid,
		PUSI:    true,
		Payload: nil,
	}

	data := make([]byte, PacketLen)
	err := pkt.ToBytes(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// //sps pps
// int len = ti->sps_len + ti->pps_len;
// len = len + 9 + 5;//pes len
// if (len > TS_PACK_LEN-4) {
//     printf("pid=%d, pes size=%d is abnormal!!!!\n", pid, len);
//     return ret;
// }
// pos ++;
// //pid
// ti->es_pid = pid;
// tmp = ti->es_pid >> 8;
// p[pos++] = 0x40 | tmp;
// tmp = ti->es_pid;
// p[pos++] = tmp;
// p[pos] = 0x10;
// int ad_len = TS_PACK_LEN - 4 - len - 1;
// if (ad_len > 0) {
//     p[pos++] = 0x30;
//     p[pos++] = ad_len;//adaptation length
//     p[pos++] = 0x00;//
//     memset(p + pos, 0xFF, ad_len-1);
//     pos += ad_len - 1;
// }else{
//     pos ++;
// }

// //pes
// p[pos++] = 0;
// p[pos++] = 0;
// p[pos++] = 1;
// p[pos++] = stream_id;
// p[pos++] = 0;//total size
// p[pos++] = 0;//total size
// p[pos++] = 0x80;//flag
// p[pos++] = 0x80;//flag
// p[pos++] = 5;//header_len
// p[pos++] = 0;//pts
// p[pos++] = 0;
// p[pos++] = 0;
// p[pos++] = 0;
// p[pos++] = 0;
// memcpy(p+pos, ti->sps, ti->sps_len);
// pos += ti->sps_len;
// memcpy(p+pos, ti->pps, ti->pps_len);
// pos += ti->pps_len;
