import { config, mount } from "@vue/test-utils";
import Message from "@/components/Message.vue";
import { expect } from "chai";
import linkifyHTML from "linkify-html";
import { createStore } from "vuex";

config.global = {
  directives: {
    Translate() {
      // do nothing in this test
    },
  },
  mixins: [
    {
      methods: {
        linkify(content) {
          return linkifyHTML(content);
        },
      },
    },
  ],
  stubs: ["FontAwesomeIcon"],
};

describe("Message.vue", () => {
  const mockStore = createStore({
    state: {
      config: {},
      currentGroup: {
        Members: [],
      },
    },
  });
  test("renders simple message without changes", () => {
    expect(Message).not.be.undefined;
    const msg = {
      message_type: "SyncMessage",
      sender: "a000000-5ddf-4fba-a6ee-2b0cb4663a6e",
      message: "🦎🍉🍓",
      timestamp: 1686505391763,
      is_outgoing: true,
      thread_id: null,
      attachments: [],
      is_sent: true,
    };

    const wrapper = mount(Message, {
      props: {
        message: msg,
        isGroup: false,
      },
      global: {
        plugins: [mockStore], //
      },
    });
    // Todo: check if message is rendered
    expect(wrapper.get('[data-test="message-text-content"]').wrapperElement.innerHTML).to.equal(
      msg.message,
    );
  });

  test("renders message with link linkified", () => {
    const expected = 'Visit <a href="http://axolotl.chat">axolotl.chat</a> if you have time';
    const msg = {
      message_type: "SyncMessage",
      sender: "a000000-5ddf-4fba-a6ee-2b0cb4663a6e",
      message: "Visit axolotl.chat if you have time",
      timestamp: 1686505391763,
      is_outgoing: true,
      thread_id: null,
      attachments: [],
      is_sent: true,
    };
    const wrapper = mount(Message, {
      props: {
        message: msg,
        isGroup: false,
      },
      global: {
        plugins: [mockStore], //
      },
    });
    // Todo: check if message is rendered

    expect(wrapper.get('[data-test="message-text-content"]').wrapperElement.innerHTML).to.equal(
      expected,
    );
  });

  test("renders message with html entities escaped", () => {
    const expected = "I <3 Axolotl";
    const msg = {
      message_type: "SyncMessage",
      sender: "a000000-5ddf-4fba-a6ee-2b0cb4663a6e",
      message: "I <3 Axolotl",
      timestamp: 1686505391763,
      is_outgoing: true,
      thread_id: null,
      attachments: [],
      is_sent: true,
    };
    const wrapper = mount(Message, {
      props: {
        message: msg,
        isGroup: false,
      },
      global: {
        plugins: [mockStore], //
      },
    });

    expect(wrapper.get('[data-test="message-text-content"]').wrapperElement.textContent).to.equal(
      expected,
    );
  });

  test("does not interpred injected html code", () => {
    const msg = {
      message_type: "SyncMessage",
      sender: "a000000-5ddf-4fba-a6ee-2b0cb4663a6e",
      message: '<div data-test="html-injection">Injected Code</div>',
      timestamp: 1686505391763,
      is_outgoing: true,
      thread_id: null,
      attachments: [],
      is_sent: true,
    };
    const wrapper = mount(Message, {
      props: {
        message: msg,
        isGroup: false,
      },
      global: {
        plugins: [mockStore], //
      },
    });
    expect(wrapper.find('[data-test="html-injection"]').exists()).to.be.false;
  });
});
