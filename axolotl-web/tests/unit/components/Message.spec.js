import { config, mount } from '@vue/test-utils'
import LinkifyHtml from 'linkifyjs/html'
import Message from '@/components/Message.vue'
import { expect } from 'chai'
import sinon from 'sinon'
import { nextTick } from 'vue';

import moment from "moment";

config.global = {
  directives: {
    Translate() {
      // do nothing in this test
    }
  },
  mixins: [
    {
      methods: {
        linkify(content) {
          return LinkifyHtml(content);
        }
      }
    }
  ],
}

function getMessage(properties) {
  return {
    ID: 'test',
    Message: '',
    Attachment: '',
    Outgoing: false,
    QuotedMessage: null,
    //ExpireTimer: 0,
    ReceivedAt: 0,
    ...properties
  }
}

describe('Message.vue', () => {
  describe('sanitization and linkify', () => {
    it('renders simple message without changes', () => {
      const msg = getMessage({
        Message: 'Test Message',
      });
      const wrapper = mount(Message, {
        props: {
          message: msg,
          isGroup: false,
          names: [ ]
        }
      });
      expect(wrapper.get('[data-test="message-text"]').wrapperElement.innerHTML, 'message html').to.equal(msg.Message)
    })

    it('renders message with link linkified', () => {
      const expected = 'Visit <a href="http://axolotl.chat">axolotl.chat</a> if you have time'
      const msg = getMessage({
        Message: 'Visit axolotl.chat if you have time',
      });
      const wrapper = mount(Message, {
        props: {
          message: msg,
          isGroup: false,
          names: [ ]
        }
      });
      expect(wrapper.get('[data-test="message-text"]').wrapperElement.innerHTML, 'message html').to.equal(expected)
    })

    it('renders message with html entities escaped', () => {
      const msg = getMessage({
        Message: 'I <3 Axolotl',
      });
      const wrapper = mount(Message, {
        props: {
          message: msg,
          isGroup: false,
          names: [ ]
        }
      });
      expect(wrapper.get('[data-test="message-text"]').wrapperElement.textContent, 'message text').to.equal(msg.Message)
    })

    it('does not interpred injected html code', () => {
      const msg = getMessage({
        Message: '<div data-test="html-injection">Injected Code</div>',
      });
      const wrapper = mount(Message, {
        props: {
          message: msg,
          isGroup: false,
          names: [ ]
        }
      })
      expect(wrapper.find('[data-test="html-injection"]').exists(), 'existence of message element').to.be.false;
    })
  })

  describe('self destroying messages', () => {
    var clock;
    beforeEach(() => {
      clock = sinon.useFakeTimers(new Date('2000-06-30T18:00:00+01:00'));

    });
    afterEach(() => {
      clock.restore();
      sinon.restore();
    })

    //recieved messages
    it('should instantly destroy recieved message beyond expire timer', () => {
      const msg = getMessage({
        ExpireTimer: 1,
        ReceivedAt: new Date('2000-06-30T17:59:58+01:00')
      });
      const $store = {
        dispatch: sinon.spy(),
      }

      const wrapper = mount(Message, {
        props: {
          message: msg,
          isGroup: false,
          names: [ ]
        },
        global: {
          mocks: {
            $store,
          }
        }
      })

      expect(wrapper.find('[data-test="message-text"]').exists(), 'existence of message element').to.be.false;
      expect($store.dispatch.calledOnce, 'dispatch called').to.be.true;
    })

    it('should destroy recieved message after reaching its expire timer', async () => {
      const msg = getMessage({
        ExpireTimer: 1,
        ReceivedAt: new Date('2000-06-30T18:00:00+01:00')
      });
      const $store = {
        dispatch: sinon.spy(),
      }

      const wrapper = mount(Message, {
        props: {
          message: msg,
          isGroup: false,
          names: [ ]
        },
        global: {
          mocks: {
            $store,
          }
        }
      })

      expect(wrapper.find('[data-test="message-text"]').exists(), 'existence of message element at first').to.be.true;
      expect($store.dispatch.notCalled, 'dispatch not called yet').to.be.true;

      clock.tick(1000)
      await nextTick();

      expect(wrapper.find('[data-test="message-text"]').exists(), 'existence of message element after timeout').to.be.false;
      expect($store.dispatch.calledOnce, 'dispatch called').to.be.true;

    })

    it('should not destroy recieved message just before reaching its expire timer', async () => {
      const msg = getMessage({
        ExpireTimer: 1,
        ReceivedAt: new Date('2000-06-30T18:00:00+01:00')
      });
      const $store = {
        dispatch: sinon.spy(),
      }

      const wrapper = mount(Message, {
        props: {
          message: msg,
          isGroup: false,
          names: [ ]
        },
        global: {
          mocks: {
            $store,
          }
        }
      })

      expect(wrapper.find('[data-test="message-text"]').exists(), 'existence of message element at first').to.be.true;
      expect($store.dispatch.notCalled, 'dispatch not called yet').to.be.true;

      clock.tick(999)
      await nextTick();

      expect(wrapper.find('[data-test="message-text"]').exists(), 'existence of message element just before timeout').to.be.true;
      expect($store.dispatch.notCalled, 'dispatch not called').to.be.true;
    })

    //send messages
    it('should instantly destroy sent message beyond expire timer', () => {
      const msg = getMessage({
        ExpireTimer: 1,
        Outgoing: true,
        SentAt: new Date('2000-06-30T17:59:58+01:00')
      });
      const $store = {
        dispatch: sinon.spy(),
      }

      const wrapper = mount(Message, {
        props: {
          message: msg,
          isGroup: false,
          names: [ ]
        },
        global: {
          mocks: {
            $store,
          }
        }
      })

      expect(wrapper.find('[data-test="message-text"]').exists(), 'existence of message element').to.be.false;
      expect($store.dispatch.calledOnce, 'dispatch called').to.be.true;
    })

    it('should destroy sent message after reaching its expire timer', async () => {
      const msg = getMessage({
        ExpireTimer: 1,
        Outgoing: true,
        SentAt: new Date('2000-06-30T18:00:00+01:00')
      });
      const $store = {
        dispatch: sinon.spy(),
      }

      const wrapper = mount(Message, {
        props: {
          message: msg,
          isGroup: false,
          names: [ ]
        },
        global: {
          mocks: {
            $store,
          }
        }
      })

      expect(wrapper.find('[data-test="message-text"]').exists(), 'existence of message element at first').to.be.true;
      expect($store.dispatch.notCalled, 'dispatch not called yet').to.be.true;

      clock.tick(1000)
      await nextTick();

      expect(wrapper.find('[data-test="message-text"]').exists(), 'existence of message element after timeout').to.be.false;
      expect($store.dispatch.calledOnce, 'dispatch called').to.be.true;

    })

    it('should not destroy sent message just before reaching its expire timer', async () => {
      const msg = getMessage({
        ExpireTimer: 1,
        Outgoing: true,
        SentAt: new Date('2000-06-30T18:00:00+01:00')
      });
      const $store = {
        dispatch: sinon.spy(),
      }

      const wrapper = mount(Message, {
        props: {
          message: msg,
          isGroup: false,
          names: [ ]
        },
        global: {
          mocks: {
            $store,
          }
        }
      })

      expect(wrapper.find('[data-test="message-text"]').exists(), 'existence of message element at first').to.be.true;
      expect($store.dispatch.notCalled, 'dispatch not called yet').to.be.true;

      clock.tick(999)
      await nextTick();

      expect(wrapper.find('[data-test="message-text"]').exists(), 'existence of message element just before timeout').to.be.true;
      expect($store.dispatch.notCalled, 'dispatch not called').to.be.true;
    })

    it('should destroy outgoing message only after it is sent', async () => {
      const msg = getMessage({
        ExpireTimer: 1,
        Outgoing: true,
      });
      const $store = {
        dispatch: sinon.spy(),
      }

      const wrapper = mount(Message, {
        props: {
          message: msg,
          isGroup: false,
          names: [ ]
        },
        global: {
          mocks: {
            $store,
          }
        }
      })

      expect(wrapper.find('[data-test="message-text"]').exists(), 'existence of message element at first').to.be.true;
      expect($store.dispatch.notCalled, 'dispatch not called yet').to.be.true;

      clock.tick(3000)
      await nextTick();

      expect(wrapper.find('[data-test="message-text"]').exists(), 'existence of message element after some time').to.be.true;
      expect($store.dispatch.notCalled, 'dispatch not called after some time').to.be.true;

      wrapper.setProps({
          message: getMessage({
            ExpireTimer: 1,
            Outgoing: true,
            SentAt: new Date(),
          })
      })

      await nextTick();

      expect(wrapper.find('[data-test="message-text"]').exists(), 'existence of message element after it was sent').to.be.true;
      expect($store.dispatch.notCalled, 'dispatch not called after it was sent').to.be.true;

      clock.tick(1000)
      await nextTick();

      expect(wrapper.find('[data-test="message-text"]').exists(), 'existence of message element after timeout passed').to.be.false;
      expect($store.dispatch.calledOnce, 'dispatch called after timeout passed').to.be.true;

    })
  })
})
