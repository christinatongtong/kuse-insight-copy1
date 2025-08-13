FROM python:3.11.13-bullseye

ENV ENVIRONMENT=cloud
COPY ./.env /.env
COPY ./requirements.txt /requirements.txt
RUN pip3 install -r /requirements.txt

COPY ./*.py /

CMD ["python3", "/main.py"]
