package apperr

type Option func(*AppError) error

func Localize(msgID string) Option {
	return func(appError *AppError) error {
		appError.optLocalizedMsgID = &msgID
		return nil
	}
}

func LocalizeInLanguage(msgID, language string) Option {
	return func(appError *AppError) error {
		appError.optLocalizedMsgID = &msgID
		appError.optLocalizedLanguage = &language
		return nil
	}
}

func LocalizeFromConfig(cfg interface{}) Option {
	return func(appError *AppError) error {
		appError.optLocalizedConfig = cfg
		return nil
	}
}
