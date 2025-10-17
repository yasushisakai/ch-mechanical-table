
run the server:

```bash
cd cmd/server
go run .
```

test using python client

```bash
. venv/bin/activate
pip install -r requirements.txt
python wsdump.py ws://localhost:8080/
```

