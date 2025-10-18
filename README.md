
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

# When the slider gets stuck

sometimes the slider gets stuck at the right side,
if that happens, the right switch to let the slider
know it's the end of the world is not working.

Usually it will work again once the switch is adjusted
slightly to the left.
