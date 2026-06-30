package kms

// ScheduleService is disabled in KCONE-integrated mode.
// All transaction submission routes through Kernel → SOR instead of the
// legacy scheduler that polled the old krypcore.com platform endpoints.
func (s *Service) ScheduleService() error {
	return nil
}
