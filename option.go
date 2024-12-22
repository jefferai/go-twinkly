// SPDX-FileCopyrightText: 2024 Jeff Mitchell <jeffrey.mitchell@gmail.com>
// SPDX-License-Identifier: APL-2.0

package twinkly

func getOpts(opt ...Option) (options, error) {
	opts := getDefaultOptions()
	for _, o := range opt {
		if o != nil {
			err := o(&opts)
			if err != nil {
				return opts, err
			}
		}
	}
	return opts, nil
}

type Option func(*options) error

type options struct {
	withHost        string
	withContentType string
}

func getDefaultOptions() options {
	return options{}
}

func WithHost(with string) Option {
	return func(o *options) error {
		o.withHost = with
		return nil
	}
}

func WithContentType(with string) Option {
	return func(o *options) error {
		o.withContentType = with
		return nil
	}
}
