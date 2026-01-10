package sys

// EnableProxy 开启系统代理
func EnableProxy(host string, port int, systemProxyMode bool) error {
	// 只有在 systemProxyMode 为 true 时才设置系统代理
	if systemProxyMode {
		_ = OnHttp(Addr{
			Host: host,
			Port: port,
		})
		_ = OnHttps(Addr{
			Host: host,
			Port: port,
		})
		_ = OnSocks(Addr{
			Host: host,
			Port: port,
		})
	}

	return nil
}

// DisableProxy 关闭代理
func DisableProxy() {
	_ = OffAll()
}
