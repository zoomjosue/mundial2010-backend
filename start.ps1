if (!(Test-Path "./frontend")) {
    git clone https://github.com/zoomjosue/mundial2010-frontend.git frontend
}

docker compose up --build