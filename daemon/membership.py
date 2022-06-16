"""Membership module provides functions for validating that all chat members are present in chats necessary by the organisation.

    Typical usage example:

    client = TelegramClient("tg-daemon", api_id, api_hash)
    await client.start()
    await membership.validate(client, config)

"""
import collections
import itertools
from collections.abc import Sequence
from typing import Dict, Iterable, List, Mapping, Set

import telethon
from telethon.tl.types import UserFull

import config

Missing = collections.namedtuple(
    "Missing", ["user_id", "missing_in_chat_id", "present_in_chat_id"])


async def validate(client: telethon.TelegramClient,
                   retg_config: config.Config):
    """Validates that the users in the chats are in the chats of their hierarchy.

    Args:
        client: Telethon Telegram client.
        config: reTGanslator bot config specifying the chats structure, etc.
    """
    chats = retg_config.all_chats()
    chat_ids = {chat.chat_id for chat in chats}
    chat_per_id = {chat.chat_id: chat for chat in chats}

    dialogs = {}

    async for dialog in client.iter_dialogs():
        if dialog.id in chat_ids:
            dialogs[dialog.id] = dialog

        if len(dialogs) == len(chat_ids):
            break

    users_per_id = {}
    user_ids_per_chat_id = {}

    for chat_id, dialog in dialogs.items():
        users_total_list = await client.get_participants(dialog)
        users = list(users_total_list)
        users_per_id.update({user.id: user for user in users})
        user_ids_per_chat_id[chat_id] = {user.id for user in users}
    missing = find_missing(retg_config, user_ids_per_chat_id)
    print_missing(users_per_id, chat_per_id, missing)


def print_missing(
    users_per_id: Mapping[int, UserFull],
    chat_per_id: Mapping[int, config.Chat],
    missing: Iterable[Missing],
):
    """Print missing memberships.

    Args:
        users_per_id: user information from Telethon.
            Used for human-readable printing.
        chat_per_id: chat configuration. Used for printing chat names.
        missing: a sequence of membership errors.
    """
    users_missing: Dict[int, Dict[int, Set[int]]] = {}

    for miss in missing:
        if miss.user_id not in users_missing:
            users_missing[miss.user_id] = {}

        if miss.missing_in_chat_id not in users_missing[miss.user_id]:
            users_missing[miss.user_id][miss.missing_in_chat_id] = set()
        users_missing[miss.user_id][miss.missing_in_chat_id].add(
            miss.present_in_chat_id)

    for user_id, chats in users_missing.items():
        user = users_per_id[user_id]
        print()
        print(f'User "{user.first_name} {user.last_name} '
              f'@{user.username} {user.phone}" is missing in:')

        for missing_chat_id, present_chat_ids in chats.items():
            present_str = ", ".join([
                f'"{chat_per_id[chat_id].aliases[0]}"'
                for chat_id in present_chat_ids
            ])
            print(
                f'\t "{chat_per_id[missing_chat_id].aliases[0]}" even though '
                f"they are in {present_str}")


def find_missing(
        retg_config: config.Config,
        user_ids_per_chat_id: Mapping[int, Set[int]]) -> List[Missing]:
    """Finds chats where required people are missing.
    Args:
        retg_config: reTGanslator config specifying the chat structure.
        user_ids_per_chat_id: specifying what users are in which chats.
    """
    lines = itertools.chain.from_iterable(
        [hierarchy_lines(chat) for chat in retg_config.chats])

    return list(
        itertools.chain.from_iterable([
            find_missing_in_hierarchy_line(user_ids_per_chat_id, line)
            for line in lines
        ]))


def hierarchy_lines(chat: config.Chat) -> List[List[int]]:
    if not chat.child_chats:
        return [[chat.chat_id]]

    child_lines = itertools.chain.from_iterable(
        [hierarchy_lines(child_chat) for child_chat in chat.child_chats])

    return [[chat.chat_id] + child_line for child_line in child_lines]


def find_missing_in_hierarchy_line(
    user_ids_per_chat_id: Mapping[int, Set[int]],
    chat_ids_line: Sequence[int],
) -> List[Missing]:
    missing_membership = []

    for i, ancestor_chat_id in enumerate(chat_ids_line):
        for descendant_chat_id in chat_ids_line[i:]:
            descendant_users = user_ids_per_chat_id[descendant_chat_id]
            ancestor_users = user_ids_per_chat_id[ancestor_chat_id]

            if not descendant_users.issubset(ancestor_users):
                missing_user_ids = descendant_users.difference(ancestor_users)
                missing_membership.extend([
                    Missing(user_id, ancestor_chat_id, descendant_chat_id)
                    for user_id in missing_user_ids
                ])

    return missing_membership
