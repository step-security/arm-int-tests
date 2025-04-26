FROM python:3.12-slim@sha256:85824326bc4ae27a1abb5bc0dd9e08847aa5fe73d8afb593b1b45b7cb4180f57

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

