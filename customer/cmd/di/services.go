package di

import "github.com/petretiandrea/beaesthetic-backend/customer/internal/application"

func (d *DiContainer) GetCustomerService() *application.CustomerService {
	return singleton(d, "customerService", func() *application.CustomerService { return application.NewCustomerService(d.GetCustomerRepository()) })
}
func (d *DiContainer) GetFidelityService() *application.FidelityService {
	return singleton(d, "fidelityService", func() *application.FidelityService { return application.NewFidelityService(d.GetFidelityRepository()) })
}
func (d *DiContainer) GetWalletService() *application.WalletService {
	return singleton(d, "walletService", func() *application.WalletService { return application.NewWalletService(d.GetWalletRepository()) })
}
