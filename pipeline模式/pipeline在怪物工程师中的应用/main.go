package service

func Usage() {
	sh := SignHandler{}
	hexes := make([]string, 1, 1)
	hexes[0] = "0x"

	//	传入触发器，请求签名ID，请求签名的方法
	for p := range sh.pipeline(hexes, sh.trigger, sh.requestSignID, sh.requestSign) {
		if p.Err != nil {
			return
		}
		return
	}
}
