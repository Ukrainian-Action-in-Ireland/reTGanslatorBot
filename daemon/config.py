import json
from typing import Iterable, Sequence


class Chat:
    chat_id: int
    aliases: Sequence[str]
    child_chats: Sequence["Chat"]

    def __init__(self, chat_id, aliases, child_chats=None):
        self.chat_id = chat_id
        self.aliases = aliases
        self.child_chats = child_chats if child_chats else []

    @staticmethod
    def from_json_dict(json_dict):
        json_dict["chat_id"] = json_dict["id"]
        del json_dict["id"]
        chat = Chat(**json_dict)
        chat.child_chats = [
            Chat.from_json_dict(child) for child in chat.child_chats
        ]

        return chat


class Config:
    chats: Sequence[Chat]
    help_contacts: Iterable[str]

    def __init__(self, chats, help_contacts):
        self.chats = chats
        self.help_contacts = help_contacts

    @staticmethod
    def from_json_dict(json_dict):
        chats = [Chat.from_json_dict(chat) for chat in json_dict["chats"]]
        config = Config(chats, json_dict["help_contacts"])

        return config

    def all_chats(self):
        res = []
        chats = [*self.chats]

        while chats:
            chat = chats.pop(0)
            chats.extend(chat.child_chats)
            res.append(chat)

        return res
