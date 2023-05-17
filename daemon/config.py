import json
from typing import Iterable, List, Sequence


class Chat:
    chat_id: int
    aliases: Sequence[str]
    members_must_be_in_any_child_chat: bool
    child_chats: Sequence["Chat"]

    def __init__(self, chat_id, aliases, members_must_be_in_any_child_chat=False, child_chats=None):
        self.chat_id = chat_id
        self.aliases = aliases
        self.members_must_be_in_any_child_chat = members_must_be_in_any_child_chat
        self.child_chats = child_chats if child_chats else []

    @staticmethod
    def from_json_dict(json_dict):
        json_dict = json_dict.copy()
        json_dict["chat_id"] = json_dict["id"]
        del json_dict["id"]
        chat = Chat(**json_dict)
        chat.child_chats = [
            Chat.from_json_dict(child) for child in chat.child_chats
        ]

        return chat


class NotificationTgChat:
    chat_id: int
    name: str

    def __init__(self, chat_id, name):
        self.chat_id = chat_id
        self.name = name

    @staticmethod
    def from_json_dict(json_dict):
        json_dict = json_dict.copy()
        json_dict["chat_id"] = json_dict["id"]
        del json_dict["id"]

        return NotificationTgChat(**json_dict)


class Notification:
    tg_chats: List[NotificationTgChat]

    def __init__(self, tg_chats=None):
        self.tg_chats = tg_chats if tg_chats else []

    @staticmethod
    def from_json_dict(json_dict):
        json_dict = json_dict.copy()
        json_dict["tg_chats"] = [
            NotificationTgChat.from_json_dict(tg_chat)
            for tg_chat in json_dict["tg_chats"]
        ]

        return Notification(**json_dict)


class MembershipValidation:
    notification: Notification

    def __init__(self, notification=None):
        self.notification = notification

    @staticmethod
    def from_json_dict(json_dict):
        json_dict = json_dict.copy()
        json_dict["notification"] = Notification.from_json_dict(
            json_dict["notification"])

        return MembershipValidation(**json_dict)


class Config:
    chats: Sequence[Chat]
    help_contacts: Iterable[str]
    membership_validation: MembershipValidation

    def __init__(self, chats, help_contacts, membership_validation):
        self.chats = chats
        self.help_contacts = help_contacts
        self.membership_validation = membership_validation

    @staticmethod
    def from_json_dict(json_dict):
        json_dict = json_dict.copy()
        json_dict["chats"] = [
            Chat.from_json_dict(chat) for chat in json_dict["chats"]
        ]
        json_dict[
            "membership_validation"] = MembershipValidation.from_json_dict(
                json_dict["membership_validation"])
        config = Config(**json_dict)

        return config

    def all_chats(self):
        res = []
        chats = [*self.chats]

        while chats:
            chat = chats.pop(0)
            chats.extend(chat.child_chats)
            res.append(chat)

        return res
