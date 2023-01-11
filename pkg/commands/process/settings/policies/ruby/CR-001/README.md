## CR-001 - Ruby - Logger Leaks

Do not send sensitive data to loggers.

Leaking sensitive data to loggers is a common cause of data leaks and can lead to data breaches. This policy looks for instances of sensitive data sent to loggers.