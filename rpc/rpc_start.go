package rpc

func (r *RPC) StopRPC(){
	r.stopHTTP()
	r.stopIPC()
	r.stopWS()
}

func (r *RPC) StartRPC() error {

	// Init rpc log
	//rpcapi.Init(node.config.DataDir, node.config.LogLevel, node.config.TestTokenHexPrivKey, node.config.TestTokenTti)

	// Start the various API endpoints, terminating all in case of errors
	if err := r.startInProcess(r.GetInProcessApis()); err != nil {
		return err
	}

	//Start rpc
	if r.config.IPCEnabled {
		if err := r.startIPC(r.GetIpcApis()); err != nil {
			r.stopInProcess()
			return err
		}
	}

	if r.config.RPCEnabled {
		apis := GetPublicApis()
		if len(r.config.PublicModules) != 0 {
			apis = GetApis(r.config.PublicModules...)
		}
		if err := r.startHTTP(r.httpEndpoint, apis, nil, r.config.HTTPCors, r.config.HttpVirtualHosts, HTTPTimeouts{}, r.config.HttpExposeAll); err != nil {
			logger.Info(err)
			r.stopInProcess()
			r.stopIPC()
			return err
		}
	}

	if r.config.WSEnabled {
		apis := GetPublicApis()
		if len(r.config.PublicModules) != 0 {
			apis =GetApis(r.config.PublicModules...)
		}
		if err := r.startWS(r.wsEndpoint, apis, nil, r.config.WSOrigins, r.config.WSExposeAll); err != nil {
			r.stopInProcess()
			r.stopIPC()
			r.stopHTTP()
			return err
		}
	}
	//if len(r.config.DashboardTargetURL) > 0 {
	//	apis := api.GetPublicApis()
	//	if len(r.config.PublicModules) != 0 {
	//		apis = api.GetApis(r.config.PublicModules...)
	//	}
	//
	//	targetUrl := r.config.DashboardTargetURL + "/ws/gvite/" + strconv.FormatUint(uint64(r.config.NetID), 10) + "@" + hex.EncodeToString(node.p2pServer.PrivateKey.PubByte())
	//
	//	u, e := url.Parse(targetUrl)
	//	if e != nil {
	//		return e
	//	}
	//	if u.Scheme != "ws" && u.Scheme != "wss" {
	//		return errors.New("DashboardTargetURL need match WebSocket Protocol.")
	//	}
	//
	//	cli, server, e := StartWSCliEndpoint(u, apis, nil, r.config.WSExposeAll)
	//	if e != nil {
	//		cli.Close()
	//		server.Stop()
	//		return e
	//	} else {
	//		r.wsCli = cli
	//	}
	//}

	return nil
}
