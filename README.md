# logfmt-formatter
This package is a streamlined fork of `TextFormatter` that provides [logfmt](https://brandur.org/logfmt) compatibility with colorization.

The default `TextFormatter` in logrus disrupts logfmt compatibility by adding colors to the output.

This package is particularly useful when working with tools like [logfmt](https://github.com/go-logfmt/logfmt).

I created this package to parse my logs using `promtail` pipelines (`decolorize` and `logfmt`).
