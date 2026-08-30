FROM python:3.12-slim

RUN useradd --system --create-home --shell /bin/bash clamavuser

EXPOSE 8000
USER clamavuser
WORKDIR /home/clamavuser

ENV PATH="/home/clamavuser/.local/bin:${PATH}"
RUN \
    python -m pip install --user setuptools
RUN \
    python -m pip install --user cvdupdate==1.1.1 && \
    cvd update

