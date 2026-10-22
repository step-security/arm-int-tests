FROM python:3.12-slim@sha256:cb33f7f771ecec2ef1225cf39b159591a040391bd73e4c01f00fdfc4c47611d2

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
