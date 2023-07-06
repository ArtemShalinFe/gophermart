FROM --platform=linux/amd64 alpine:latest

RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

COPY ../cmd/accrual/accrual_linux_amd64 /

CMD /accrual_linux_amd64 -a ${ACCRUAL_ADDRESS} -d ${DSN}