### Deploy webhook
1. Make sure you have `gcloud` installed and executed `gcloud auth login`
2. Put the correct configuration in the `config.json`
3. `export BOT_TOKEN=<your_bot_token>`
4. `make deploy`

### Run membership validation
```shell
export TG_API_ID=165292; export TG_API_HASH=940c7531dccfff4876cda02d52fe6771503b8fb57b; python daemon/main.py
```
