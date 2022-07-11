import asyncio
import json
import os
import sys

from telethon import TelegramClient, events, utils

import membership
from config import Config


def validate_env():
    for env in ["TG_API_ID", "TG_API_HASH"]:
        if env not in os.environ:
            raise ValueError(
                f"Env variable {env} must be specified. Env: {os.environ}")


async def run_services(api_id, api_hash, config):
    client = TelegramClient("tg-daemon", api_id, api_hash)
    await client.start()
    await membership.validate(client, config)


if __name__ == "__main__":
    config_file_name = "config.json"

    if len(sys.argv) > 1:
        config_file_name = sys.argv[1]

    validate_env()
    TG_API_ID = os.environ["TG_API_ID"]
    TG_API_HASH = os.environ["TG_API_HASH"]

    with open(config_file_name) as f:
        config = Config.from_json_dict(json.load(f))

    asyncio.run(run_services(TG_API_ID, TG_API_HASH, config))
