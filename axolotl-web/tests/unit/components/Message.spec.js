import { expect } from 'chai'
import LinkifyHtml from 'linkifyjs/html'
import Message from '@/components/Message.vue'
import { shallowMount } from '@vue/test-utils'

const wrapperConfig = {
  global: {
    directives: {
      Translate: function () { }
    }
  },
  methods: {
    linkify: function (content) {
      return LinkifyHtml(content);
    },
  }
}

describe('Message.vue', () => {
  it('renders simple message without changes', () => {
    const msg = {
        ID: 'test',
        Message: 'Test Message',
        Attachment: '',
        Outgoing: false,
        QuotedMessage: null,
        ExpireTimer: 0
    }
    const wrapper = shallowMount(Message, {
      ...wrapperConfig,
      props: {
        message: msg,
        isGroup: false,
        names: [ ]
      }
    })
    console.log(wrapper.html())
    expect(wrapper.find('.message-text-content').innerHTML).toMatch(msg.Message)
  })

  it('renders message with link linkified', () => {
    const expected = 'Visit <a href="axolotl.chat">axolotl.chat</a> if you have time'
    const msg = {
        ID: 'test',
        Message: 'Visit axolotl.chat if you have time',
        Attachment: '',
        Outgoing: false,
        QuotedMessage: null,
        ExpireTimer: 0
    }
    const wrapper = shallowMount(Message, {
      ...wrapperConfig,
      props: {
        message: msg,
        isGroup: false,
        names: [ ]
      }
    })
    expect(wrapper.find('.message-text-content').innerHTML).toMatch(expected)
  })

  it('renders message with html entities escaped', () => {
    const expected = 'I &lt;3 Axolotl'
    const msg = {
        ID: 'test',
        Message: 'I <3 Axolotl',
        Attachment: '',
        Outgoing: false,
        QuotedMessage: null,
        ExpireTimer: 0
    }
    const wrapper = shallowMount(Message, {
      ...wrapperConfig,
      props: {
        message: msg,
        isGroup: false,
        names: [ ]
      }
    })
    expect(wrapper.find('.message-text-content').innerHTML).toMatch(expected)
  })
})
