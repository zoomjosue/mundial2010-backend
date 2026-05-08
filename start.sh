#!/bin/bash

if [ ! -d "frontend" ]; then
  git clone https://github.com/zoomjosue/mundial2010-frontend.git frontend
fi

docker compose up --build