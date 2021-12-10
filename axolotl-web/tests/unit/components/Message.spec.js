import { config, mount } from '@vue/test-utils'
import LinkifyHtml from 'linkifyjs/html'
import Message from '@/components/Message.vue'
import MockDate from 'mockdate'
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

describe('Message.vue', () => {
  describe('sanitization and linkify', () => {
    it('renders simple message without changes', () => {
      const msg = {
          ID: 'test',
          Message: 'Test Message',
          Attachment: '',
          Outgoing: false,
          QuotedMessage: null,
          ExpireTimer: 0,
          ReceivedAt: 0,
      }
      const wrapper = mount(Message, {
        props: {
          message: msg,
          isGroup: false,
          names: [ ]
        }
      })
      expect(wrapper.get('[data-test="message-text"]').wrapperElement.innerHTML, 'message html').to.equal(msg.Message)
    })

    it('renders message with link linkified', () => {
      const expected = 'Visit <a href="http://axolotl.chat">axolotl.chat</a> if you have time'
      const msg = {
          ID: 'test',
          Message: 'Visit axolotl.chat if you have time',
          Attachment: '',
          Outgoing: false,
          QuotedMessage: null,
          ExpireTimer: 0
      }
      const wrapper = mount(Message, {
        props: {
          message: msg,
          isGroup: false,
          names: [ ]
        }
      })
      expect(wrapper.get('[data-test="message-text"]').wrapperElement.innerHTML, 'message html').to.equal(expected)
    })

    it('renders message with html entities escaped', () => {
      const expected = 'I <3 Axolotl'
      const msg = {
          ID: 'test',
          Message: 'I <3 Axolotl',
          Attachment: '',
          Outgoing: false,
          QuotedMessage: null,
          ExpireTimer: 0
      }
      const wrapper = mount(Message, {
        props: {
          message: msg,
          isGroup: false,
          names: [ ]
        }
      })
      expect(wrapper.get('[data-test="message-text"]').wrapperElement.textContent, 'message text').to.equal(expected)
    })

    it('does not interpred injected html code', () => {
      const msg = {
          ID: 'test',
          Message: '<div data-test="html-injection">Injected Code</div>',
          Attachment: '',
          Outgoing: false,
          QuotedMessage: null,
          ExpireTimer: 0
      }
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
    beforeEach(() => {
      MockDate.set('2000-06-30T18:00:00+01:00');

    });
    afterEach(() => {
      MockDate.reset();
      sinon.restore();
    })
    it('should instantly destroy message beyond expire timer', () => {
      const msg = {
          ID: 'test',
          Message: '<div data-test="html-injection">Injected Code</div>',
          Attachment: '',
          Outgoing: false,
          QuotedMessage: null,
          ExpireTimer: 1,
          ReceivedAt: new Date('2000-06-30T17:59:58+01:00')
      }
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

    it('should destroy message after reaching its expire timer', async () => {
      const msg = {
          ID: 'test',
          Message: '<div data-test="html-injection">Injected Code</div>',
          Attachment: '',
          Outgoing: false,
          QuotedMessage: null,
          ExpireTimer: 1,
          ReceivedAt: new Date('2000-06-30T18:00:00+01:00')
      }
      const $store = {
        dispatch: sinon.spy(),
      }
      var clock = sinon.useFakeTimers()

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

      clock.tick(2100)
      await nextTick();

      expect(wrapper.find('[data-test="message-text"]').exists(), 'existence of message element after timeout').to.be.false;
      expect($store.dispatch.calledOnce, 'dispatch called').to.be.true;
      clock.restore();
    })
  })
})
