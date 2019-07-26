# kit-mw

A set of reusable middlewares for use with [go-kit](https://github.com/go-kit/kit). 

# How

Go-kit provides a useful abstractions that allow for easy wrapping of handlers using middlewares.
You can read ore about this [here](https://gokit.io/faq/#middlewares-mdash-what-are-middlewares-in-go-kit).

# Middlewares

* [eplogger](eplogger) - An endpoint middleware that provides logging of every request with request and response type-specific fields 
logged based upon the implementation of the `AppendKeyvalser` interface.