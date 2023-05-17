import unittest

from config import Chat, Config, MembershipValidation
from membership import find_missing_in_children, hierarchy_lines, Missing
from parameterized import parameterized


class Test_hierarchy_lines(unittest.TestCase):

    def test_chat_with_no_children(self):
        self.assertEqual(
            hierarchy_lines(Chat(123, "First")),
            [[123]],
            "Should be a single line consisting of the chat",
        )

    def test_chat_with_two_children(self):
        self.assertEqual(
            hierarchy_lines(
                Chat(
                    123,
                    "First",
                    child_chats=[
                        Chat(234, "Second"),
                        Chat(345, "Third"),
                    ],
                )),
            [
                [123, 234],
                [123, 345],
            ],
            "Should be two lines: one for each child",
        )


class Test_find_missing_in_children(unittest.TestCase):
    @parameterized.expand([
        (
            "no chats return no missing",
            Config(chats=[], help_contacts=[],
                   membership_validation=MembershipValidation()),
            {},
            [],
        ),
        (
            "single chat returns no missing",
            Config(chats=[Chat(1234, [], members_must_be_in_any_child_chat=True)
                          ], help_contacts=[], membership_validation=MembershipValidation()),
            {1234: []},
            [],
        ),
        (
            "user not present in any child chat is returned",
            Config(chats=[Chat(1234, [], members_must_be_in_any_child_chat=True, child_chats=[
                Chat(4321, []),
                Chat(2222, []),
            ])
            ], help_contacts=[], membership_validation=MembershipValidation()),
            {
                1234: [11111, 2424],
                4321: [],
                2222: [2424],
            },
            [
                Missing(11111, 1234, missing_in_children_chats=True),
            ],
        ),
        (
            "members_must_be_in_any_child_chat=False returns no missing",
            Config(chats=[Chat(1234, [], child_chats=[
                Chat(4321, []),
                Chat(2222, []),
            ])
            ], help_contacts=[], membership_validation=MembershipValidation()),
            {
                1234: [11111, 2222],
                4321: [],
                2222: [],
            },
            [],
        ),
        (
            "child chat having members_must_be_in_any_child_chat=True but parent chat not having it returns no missing",
            Config(chats=[Chat(1234, [], child_chats=[
                Chat(4321, [], members_must_be_in_any_child_chat=True,
                     child_chats=[Chat(2222, [])]),
            ])
            ], help_contacts=[], membership_validation=MembershipValidation()),
            {
                1234: [],
                4321: [2424],
                2222: [],
            },
            [],
        ),
        (
            "child chats having users not present in the parent chat do not return missing in child chats",
            Config(
                chats=[Chat(1234, [], members_must_be_in_any_child_chat=True,
                            child_chats=[Chat(4321, [])])],
                help_contacts=[], membership_validation=MembershipValidation()),
            {
                1234: [],
                4321: [2424],
            },
            [],
        ),
    ])
    def test_returned_missing_is_correct(self, _, config, user_ids_per_chat_id, want_missing):
        got_missing = find_missing_in_children(config, user_ids_per_chat_id)

        self.assertEqual(want_missing, list(got_missing))


if __name__ == "__main__":
    unittest.main()
