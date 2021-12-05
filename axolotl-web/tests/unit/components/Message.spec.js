import { expect } from 'chai'
import { shallowMount } from '@vue/test-utils'

import Message from '@/components/Message.vue'

describe('Message.vue', () => {
  it('renders simple message without changes', () => {
    const msg = {
        ID: 'test',
        Message: 'Test Message',
        Attachment: [],
        Outgoing: false,
        QuotedMessage: null
    }
    const wrapper = shallowMount(Message, {
      props: {
        message: msg,
        isGroup: false,
        names: [ ]
      }
    })
    expect(wrapper.findComponent('.message-text-content').html()).toMatch(msg.Message)
  })
})
