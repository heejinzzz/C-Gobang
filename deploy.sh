if mkdir -p /data/C-Gobang; then
  if cp -r ./dockerVolumes /data/C-Gobang/dockerVolumes; then
    docker compose up -d
  else
    echo "Fail to execute \"cp -r ./dockerVolumes /data/C-Gobang/dockerVolumes\" !"
    exit 8
  fi
else
  echo "Fail to execute \"mkdir -p /data/C-Gobang\" !"
  exit 9
fi
