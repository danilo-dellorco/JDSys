import os;
os.system("sudo cp /home/ec2-user/progetto-sdcc/src/server/progettosdcc.service /etc/systemd/system")
os.system("sudo systemctl daemon-reload")
os.system("sudo systemctl start progettosdcc.service")
os.system("sudo systemctl status progettosdcc.service -l")