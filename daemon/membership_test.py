import unittest

from config import Chat
from membership import hierarchy_lines


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


if __name__ == "__main__":
    unittest.main()
