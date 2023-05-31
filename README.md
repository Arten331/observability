# Observability

These packages are designed to provide observability features for personal use.

## Packages

### _logger_

The logger package is a wrapper built on top of the zap.Logger library. It provides an easy-to-use interface for logging within your GOLANG applications. The logger package offers various logging levels, including debug, info, warning, and error, allowing you to effectively manage and track application logs.
Tracer Package

### _tracer_

The tracer package is a wrapper for working with application traces. It utilizes the go.opentelemetry.io/otel library and is specifically designed to initialize and export spans to Jaeger. With the tracer package, you can easily instrument your GOLANG applications to capture distributed traces, enabling effective monitoring and debugging.
Metrics Package

### _metrics_
 
Wrapper for registering metrics within your GOLANG applications. It provides guidelines on how services within your application should register and expose metrics. By using this package, you can collect and analyze key performance indicators (KPIs) to gain insights into the behavior and efficiency of your application.


## Personal Use

Please note that this repository and its packages are intended for personal use only. While they can provide valuable observability capabilities for your GOLANG applications, they may not be suitable for production environments or large-scale deployments. Use them at your own discretion.

## Contributions

Contributions to this repository are not accepted at the moment, as it is meant for personal use only.

## License

This repository and its packages are provided under the [MIT License](LICENSE). Feel free to modify and adapt the code for your personal use.