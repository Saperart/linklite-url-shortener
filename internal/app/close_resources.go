package app

func (a *App) closeResources() {
	a.closeOnce.Do(func() {
		if a.pool != nil {
			a.pool.Close()
			a.logger.Info("postgres pool closed")
		}
	})
}
