 <template>
   <div :class="{'col-12':true,
               'outgoing': message.Outgoing,
               'sent':message.IsSent && message.Outgoing,
               'read':message.IsRead && message.Outgoing,
               'delivered':message.Receipt && message.Outgoing,
               'incoming':!message.Outgoing,
               'status':message.Flags>0&&message.Flags!=11&&message.Flags!=13&&message.Flags!=14
               ||message.StatusMessage||message.Attachment.includes('null')&&message.Message=='',
               'error':message.SentAt==0||message.SendingError
               }"
       v-bind:key="message.ID">
     <div class="message" v-if="verifySelfDestruction(message)">
       <div class="sender" v-if="!message.Outgoing&&isGroup&&message.Flags==0">
         <div v-if="names[message.Source]">
           {{names[message.Source]}}
         </div>
         <div v-else>{{getName(message.Source)}}</div>
       </div>
       <div v-if="message.Attachment!=''" class="attachment">
         <div class="gallery" v-if="isAttachmentArray(message.Attachment)">
           <div  v-for="m in isAttachmentArray(message.Attachment)"
             v-bind:key="m.File">
             <div v-if="m.CType==2" class="attachment-img">
               <img  :src="'http://localhost:9080/attachments?file='+m.File" @click="$emit('showFullscreenImg',m.File)"/>
             </div>
             <div v-else-if="m.CType==3" class="attachment-audio">
               <audio controls>
                 <source :src="'http://localhost:9080/attachments?file='+m.File" type="audio/mpeg">
                   <span v-translate>Your browser does not support the audio element.</span>
               </audio>
             </div>
             <div v-else-if="m.File!='' &&m.CType==0" class="attachment-file">
               <a @click="shareAttachment(m.File,$event)" :href="'http://localhost:9080/attachments?file='+m.File">{{m.FileName?m.FileName:m.File}}</a>
             </div>
             <div v-else-if="m.CType==5" class="attachment-video" @click="$emit('showFullscreenVideo', m.File)">
               <video>
                 <source :src="'http://localhost:9080/attachments?file='+m.File">
                   <span v-translate>Your browser does not support the audio element.</span>
               </video>
             </div>
             <div v-else-if="m.File!=''" class="attachment">
               <span v-translate>Not supported mime type:</span> {{m.CType}}
             </div>
           </div>
         </div>
         <!-- this is legacy code -->
         <div v-else-if="message.CType==2" class="attachment-img">
           <img  :src="'http://localhost:9080/attachments?file='+message.Attachment" @click="$emit('showFullscreenImg', message.Attachment)"/>
         </div>
         <div v-else-if="message.CType==3" class="attachment-audio">
           <audio controls>
             <source :src="'http://localhost:9080/attachments?file='+message.Attachment" type="audio/mpeg">
               <span v-translate>Your browser does not support the audio element.</span>
           </audio>
         </div>
         <div v-else-if="message.Attachment!='null'&&message.CType==0" class="attachment-file">
           {{message.Attachment}}
           <a :href="'http://localhost:9080/attachments?file='+message.Attachment">File</a>
         </div>
         <div v-else-if="message.CType==5" class="attachment-video" @click="$emit('showFullscreenVideo', message.Attachment)">
           <video>
             <source :src="'http://localhost:9080/attachments?file='+message.Attachment">
               <span v-translate>Your browser does not support the video element.</span>
           </video>
         </div>

         <div v-else-if="message.Attachment!='null'" class="attachment">
           <span v-translate>Not supported mime type:</span> {{message.CType}}
         </div>
       </div>
       <div class="message-text">
         <blockquote v-if="message.QuotedMessage != null">
           <cite v-if="message.QuotedMessage.Outgoing"  v-translate>You</cite>
           <cite v-else>{{getName(message.QuotedMessage.Source)}}</cite>
           <p>{{message.QuotedMessage.Message}}</p>
         </blockquote>
         <div class="message-text-content" v-html="message.Message" v-linkified ></div>
         <div class="status-message" v-if="message.Attachment.includes('null')&&message.Message==''&&message.Flags==0">
           <span v-translate>Set timer for self-destructing messages </span>
           <div> {{humanifyTimePeriod(message.ExpireTimer)}}</div>
         </div>
         <div v-if="message.Flags==10" v-translate>Unsupported message type: sticker</div>
       </div>
       <div class="meta" v-if="message.SentAt!=0">
         <div class="time">{{humanifyDate(message.SentAt)}}</div>
         <div v-if="message.ExpireTimer>0&&message.Message!=''">
           <div class="circle-wrap">
             <div class="circle">
               <div class="mask full" :style="'transform: rotate('+timerPercentage(message)+'deg)'">
                 <div class="fill" :style="'transform: rotate('+timerPercentage(message)+'deg)'"></div>
               </div>
               <div class="mask half">
                 <div class="fill" :style="'transform: rotate('+timerPercentage(message)+'deg)'"></div>
               </div>
               <div class="inside-circle">
               </div>
             </div>
           </div>
         </div>
         <div v-if="message.Outgoing" class="transfer-indicator"></div>
       </div>
       <div v-else class="col-12 meta">
         Error
       </div>
     </div>
   </div>
</template>

<script>
import moment from 'moment';
import { mapState } from 'vuex';

export default {
  name: 'Message',
  props: ['message', 'isGroup', 'names'],
  computed: mapState(['contacts']),
  methods: {
    getName(tel){
      if(this.contacts!=null){
        if(typeof this.names[tel]=="undefined"){
          var contact = this.contacts.find(function(element) {
            return element.Tel == tel;
          });
          if(typeof contact!="undefined"){
            this.names[tel]=contact.Name;
            return contact.Name
          }else{
            this.names[tel] = tel;
            return tel
          }
        }else return this.names[tel]
      }
      return tel;
    },
    isAttachmentArray(input){
      try{
        var attachments = JSON.parse(input)
        return attachments;

      } catch(e){
        return false;
      }
      // JSON.parse(input)
    },
    verifySelfDestruction(m){
      if(m.ExpireTimer!=0){
        // console.log(m.ExpireTimer,m.SentAt, m.ReceivedAt, Date.now())
        if(m.ReceivedAt!=0){
          // hide destructed messages but not timer settings
          var r = moment(m.ReceivedAt)
          var duration = moment.duration(r.diff(moment.now()));
          if((duration.asSeconds()+m.ExpireTimer)<0&&m.Message!=""){
            return false;
          }
        }
        else if (m.SentAt!=0){
          var rS = moment(m.SentAt)
          var durationS = moment.duration(rS.diff(moment.now()));
          if((durationS.asSeconds()+m.ExpireTimer)<0&&m.Message!=""){
            return false;
          }
        }
      }
      return true
    },
    humanifyDate(inputDate){
      moment.locale(this.$language.current);
      var date = new moment(inputDate);
      var min = moment().diff(date, 'minutes')
      if(min<60){
        if(min == 0) return "now"
        return moment().diff(date, 'minutes') +" min"
      }
      var hours = moment().diff(date, 'hours')
      if(hours <24) return hours + " h"
      return date.format("DD. MMM");
    },
    humanifyTimePeriod(time){
      if(time<60)
        return time +" s";
      else if(time<60*60)
        return time/60+" m"
      else if(time<60*60*24)
        return time/60/60+" h"
      else if(time<60*60*24)
        return time/60/60/24+" d"
      return time
    }
  }
}
</script>

<style lang="scss" scoped>
.message-text{
  overflow-wrap: break-word;
}
.incoming {
  text-align:left;
}
.outgoing{
  display:flex;
  justify-content:flex-end;
}
.meta {
  display: flex;
  align-items: center;
  font-size: 11px;
}
.outgoing .meta {
  justify-content: flex-end;
}
.message {
  margin-bottom:10px;
  padding: 8px 12px;
  border-radius: 10px;
  max-width:85%;
  text-align:left;
  min-width:100px;
}
.error .message{
  border: solid #f7663a 2px;
}
.sender{
  font-size:0.9rem;
  font-weight:bold;
}
video,
.attachment-img img {
    max-width: 100%;
    max-height: 80vh;
}
.status .message {
  background-color:transparent;
  width:100%;
  font-weight:600;
  text-align: center;
}
.status .status-message{
  width:100%;
  display:flex;
  justify-content:center;
  font-weight:600;
  text-align: center;
  flex-direction:column;
}
.status .status-message span{
  padding-right:4px;
}
.status .meta{
  text-align: center;
  justify-content:center;
}
.transfer-indicator {
  width: 18px;
  height: 12px;
  margin-left: 4px;
  background-repeat: no-repeat;
  background-position: left center;
}
.error .transfer-indicator {
  background-image: url("../assets/images/warning.svg");
}
.circle-wrap {
  margin-top: 3px;
  margin-left: 5px;
  width: 15px;
  height: 15px;
  background: #e6e2e7;
  border-radius: 50%;
}
.circle-wrap .circle .mask,
.circle-wrap .circle .fill {
  width: 16px;
  height: 16px;
  position: absolute;
  border-radius: 50%;
}
.circle-wrap .circle .mask {
  clip: rect(0px, 16px, 16px, 8px);
}
.circle-wrap .circle .mask .fill {
  clip: rect(0px, 8px, 16px, 0px);
  background-color: #9e00b1;
}

.circle-wrap .circle.p0 .mask.full,
.circle-wrap .circle.p0 .fill {
  /* animation: fill ease-in-out 3s; */
  transform: rotate(00deg);
}
.circle-wrap .circle.p50 .mask.full,
.circle-wrap .circle.p50 .fill {
  /* animation: fill ease-in-out 3s; */
  transform: rotate(180deg);
}
.circle-wrap .circle.p100 .mask.full,
.circle-wrap .circle.p100 .fill {
  /* animation: fill ease-in-out 3s; */
  transform: rotate(360deg);
}
.message-text .message-text-content{
  white-space: pre-line;
}
.gallery{
  display:flex;
}
.gallery img{
  padding-right:3px;
  padding-bottom:3px;
}
blockquote {
  padding: 0.5rem;
  margin-bottom: 5px;
  background-color: #00000044;
  border-left: solid 4px #00000044;
  border-radius: 4px;

  cite {
    font-style: normal;
    font-weight: bold;
  }

  p {
    margin: 0;
  }
}
</style>
