#### Note:


### Message sending states
1. Click sending a message ->
2. Message isn't shown until backend processed it ->
3.1 Message sent to signal successfully (SendingError = false, isSent = false, isRead=false)
3.2 Messege sending failure because of missing connection for example  (SendingError = true, isSent = false, isRead=false)
4. Message was received by other person (SendingError = false, isSent = true, isRead=false)
5. Message was read by other person (SendingError = false, isSent = true, isRead=true)
