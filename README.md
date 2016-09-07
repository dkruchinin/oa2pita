## Preamble
OAuth2 authorization is pain in the ass, especially if you have to deal with it from a command line app. When I had to write a CLI tool that used oauth2 some time ago, I couldn't find many examples of how people usually deal with the fact that there's an authorization link a user is supposed to follow that eventually redirects the user to a redirect URL specified earlier. It all doesn't quite fit to what I imagine a good CLI experience is.

## The app
oa2pita demostrates how to deal with OAuth2 from a CLI tool (basically by launching up an http server and dealing with the data you get from an oauth2 server). Since it's just an example I didn't bother making it customisable. I picked bitbucket as my testing animal (mostly because its API documentation sucks) but other services should work in exactly (or almost exactly) the same way.

So, what oa2pita can do? Not too much. When you launch it for the first time, it tries to get a new token from bitbucket, saves it to a file ($HOME/.bb.token) and dumps it out to your terminal. On subsequent runs it'll just output the saved token (it's brilliant at printing out things) up until the point when the token gets expired in which case oa2pita tries to refresh it.

## Setup

Make sure your $GOPATH/bin directory is in your $PATH

    go get -u github.com/dkruchinin/oa2pita

Go to bitbucket, 
1. Click on your profile picture (top right corner if they haven't moved it around)
2. -> bitbucket settings
3. -> OAuth
4. -> Find and press "Add Consumer" button
5. Give the consumer a name (how about "pita"?) 
6. Enter the callback URL and set up the permissions in exactly the same way you see it in the picture below.

![bitbucket](https://github.com/dkruchinin/oa2pita/blob/master/img/bb.png)

Now when you click on the consumer, you'll see generated *Key* and *Secret*. You'll need them

## Usage

    oa2pita
    Running auth server ...
    No token file exists at the moment. Trying to get a new token...
    Please enter a clinent ID: <put here your bibucket key and press enter>
    And a client secret: <and put your secret over here>
    To authorize with bitbucket please follow this url:      https://bitbucket.org/site/oauth2/authorize?client_id=<KEY>&response_type=code&scope=project
    ...
