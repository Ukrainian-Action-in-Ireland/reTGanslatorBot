import asyncio
import json
import os
import sys

from telethon import TelegramClient, events, utils

import membership
from config import Config

USAGE = f"{sys.argv[0]} [path/to/bot_config_file.json] [path/to/tg_session_file]"


def validate_env():
    for env in ["TG_API_ID", "TG_API_HASH"]:
        if env not in os.environ:
            raise ValueError(
                f"Env variable {env} must be specified. Env: {os.environ}")


async def run_services(api_id, api_hash, config, tg_session_file_name):
    client = TelegramClient(tg_session_file_name, api_id, api_hash)
    await client.start()
    await membership.validate(client, config)


if __name__ == "__main__":
    config_file_name = "config.json"
    tg_session_file_name = "tg-daemon"

    if len(sys.argv) > 3:
        print(f"Got more than 2 arguments: {sys.argv}\n {USAGE}")
        exit()
    elif len(sys.argv) == 3:
        config_file_name = sys.argv[1]
        tg_session_file_name = sys.argv[2]
    elif len(sys.argv) == 2:
        print(f"Got 1 argument. Expected 0 or 2. Got: {sys.argv}\n {USAGE}")
        exit()

    validate_env()
    TG_API_ID = os.environ["TG_API_ID"]
    TG_API_HASH = os.environ["TG_API_HASH"]

    with open(config_file_name) as f:
        config = Config.from_json_dict(json.load(f))

    asyncio.run(
        run_services(TG_API_ID, TG_API_HASH, config, tg_session_file_name))
