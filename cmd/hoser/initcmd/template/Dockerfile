# syntax=docker/dockerfile:1

FROM hoser:latest

FROM python:3-alpine

WORKDIR /app

COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

COPY --from=hoser:latest /usr/bin/hoser /usr/bin/hoser

COPY . ./

RUN ./.hoser/bootstrap.sh

CMD [ "sh", "-c", "exec hoser run *.hos"]