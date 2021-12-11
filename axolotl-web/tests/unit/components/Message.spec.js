import { config, mount } from '@vue/test-utils'
import LinkifyHtml from 'linkifyjs/html'
import Message from '@/components/Message.vue'
import chai from 'chai'
import { nextTick } from 'vue';
import sinon from 'sinon'
import sinonChai from 'sinon-chai'

const { expect } = chai

chai.use(sinonChai);

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

/**
 * Helper to initialise message objects, setting some default values.
 * @param properties adds or overwrites properties of the generated message.
 */
function getMessage(properties) {
  return {
    ID: 'test',
    Message: '',
    Attachment: '',
    Outgoing: false,
    QuotedMessage: null,
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
    let clock = null;
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
      expect($store.dispatch, 'dispatch called').to.have.been.calledOnce;
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
      expect($store.dispatch, 'dispatch not called yet').not.to.have.been.called;

      clock.tick(1000)
      await nextTick();

      expect(wrapper.find('[data-test="message-text"]').exists(), 'existence of message element after timeout').to.be.false;
      expect($store.dispatch, 'dispatch called').to.have.been.calledOnce;

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
      expect($store.dispatch, 'dispatch not called yet').not.to.have.been.called;

      clock.tick(999)
      await nextTick();

      expect(wrapper.find('[data-test="message-text"]').exists(), 'existence of message element just before timeout').to.be.true;
      expect($store.dispatch, 'dispatch not called').not.to.have.been.called;
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
      expect($store.dispatch, 'dispatch called').to.have.been.calledOnce;
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
      expect($store.dispatch, 'dispatch not called yet').not.to.have.been.called;

      clock.tick(1000)
      await nextTick();

      expect(wrapper.find('[data-test="message-text"]').exists(), 'existence of message element after timeout').to.be.false;
      expect($store.dispatch, 'dispatch called').to.have.been.calledOnce;

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
      expect($store.dispatch, 'dispatch not called yet').not.to.have.been.called;

      clock.tick(999)
      await nextTick();

      expect(wrapper.find('[data-test="message-text"]').exists(), 'existence of message element just before timeout').to.be.true;
      expect($store.dispatch, 'dispatch not called').not.to.have.been.called;
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
      expect($store.dispatch, 'dispatch not called yet').not.to.have.been.called;

      clock.tick(3000)
      await nextTick();

      expect(wrapper.find('[data-test="message-text"]').exists(), 'existence of message element after some time').to.be.true;
      expect($store.dispatch, 'dispatch not called after some time').not.to.have.been.called;

      wrapper.setProps({
          message: getMessage({
            ExpireTimer: 1,
            Outgoing: true,
            SentAt: new Date(),
          })
      })

      await nextTick();

      expect(wrapper.find('[data-test="message-text"]').exists(), 'existence of message element after it was sent').to.be.true;
      expect($store.dispatch, 'dispatch not called after it was sent').not.to.have.been.called;

      clock.tick(1000)
      await nextTick();

      expect(wrapper.find('[data-test="message-text"]').exists(), 'existence of message element after timeout passed').to.be.false;
      expect($store.dispatch, 'dispatch called after timeout passed').to.have.been.calledOnce;

    })
  })
})
