const Alexa = require("ask-sdk-core");
// Replace with http (both) if you want to use unsecure protocol
const https = require("https");

// Host of the her server to call
const homeHost = "alexa.mydomain.com";
// Path of the URL to reach her (use "/" if none)
const homePath = "/alexa/";

function callHome(data) {
    // TODO: auth https://developer.amazon.com/docs/account-linking/add-account-linking-logic-custom-skill.html
    const options = {
        host: homeHost,
        path: homePath,
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "Content-Length": data.length
        }
    };

    const req = https.request(options, res => {
        res.setEncoding("utf8");
        var responseString = "";

        //accept incoming data asynchronously
        res.on("data", chunk => {
            responseString = responseString + chunk;
        });

        //return the data when streaming is complete
        res.on("end", () => {
            console.log(responseString);
        });
    });
    req.on("error", error => {
        console.error(error);
    });
    req.write(data);
    req.end();
}

const SwitchOffTheLightHandler = {
    canHandle(handlerInput) {
        return (
            handlerInput.requestEnvelope.request.type === "IntentRequest" &&
            handlerInput.requestEnvelope.request.intent.name ===
                "SwitchOffTheLight"
        );
    },
    handle(handlerInput) {
        let speakOutput = "";
        const room =
            handlerInput.requestEnvelope.request.intent.slots.room.value;
        if (room === null || room === "") {
            speakOutput = "Non ho capito cosa devo spengere";
        } else {
            speakOutput = `Ok, spengo ${room}`;
            const data = JSON.stringify({
                action: "switch-off",
                room
            });
            callHome(data);
        }
        return (
            handlerInput.responseBuilder
                .speak(speakOutput)
                //.reprompt('add a reprompt if you want to keep the session open for the user to respond')
                .getResponse()
        );
    }
};
const SwitchOnTheLightHandler = {
    canHandle(handlerInput) {
        return (
            handlerInput.requestEnvelope.request.type === "IntentRequest" &&
            handlerInput.requestEnvelope.request.intent.name ===
                "SwitchOnTheLight"
        );
    },
    handle(handlerInput) {
        let speakOutput = "";
        const room =
            handlerInput.requestEnvelope.request.intent.slots.room.value;
        if (room === null || room === "") {
            speakOutput = "Non ho capito cosa devo accendere";
        } else {
            speakOutput = `Ok, accendo ${room}`;
            const data = JSON.stringify({
                action: "switch-on",
                room
            });
            callHome(data);
        }
        return (
            handlerInput.responseBuilder
                .speak(speakOutput)
                //.reprompt('add a reprompt if you want to keep the session open for the user to respond')
                .getResponse()
        );
    }
};
const LaunchRequestHandler = {
    canHandle(handlerInput) {
        return (
            Alexa.getRequestType(handlerInput.requestEnvelope) ===
            "LaunchRequest"
        );
    },
    handle(handlerInput) {
        const speakOutput =
            "Welcome, you can say Hello or Help. Which would you like to try?";
        return handlerInput.responseBuilder
            .speak(speakOutput)
            .reprompt(speakOutput)
            .getResponse();
    }
};
const HelpIntentHandler = {
    canHandle(handlerInput) {
        return (
            Alexa.getRequestType(handlerInput.requestEnvelope) ===
                "IntentRequest" &&
            Alexa.getIntentName(handlerInput.requestEnvelope) ===
                "AMAZON.HelpIntent"
        );
    },
    handle(handlerInput) {
        const speakOutput =
            "Puoi usarmi per accendere la luce. Cosa devo accendere?";

        return handlerInput.responseBuilder
            .speak(speakOutput)
            .reprompt(speakOutput)
            .getResponse();
    }
};
const CancelAndStopIntentHandler = {
    canHandle(handlerInput) {
        return (
            Alexa.getRequestType(handlerInput.requestEnvelope) ===
                "IntentRequest" &&
            (Alexa.getIntentName(handlerInput.requestEnvelope) ===
                "AMAZON.CancelIntent" ||
                Alexa.getIntentName(handlerInput.requestEnvelope) ===
                    "AMAZON.StopIntent")
        );
    },
    handle(handlerInput) {
        const speakOutput = "Ciao!";
        return handlerInput.responseBuilder.speak(speakOutput).getResponse();
    }
};
const SessionEndedRequestHandler = {
    canHandle(handlerInput) {
        return (
            Alexa.getRequestType(handlerInput.requestEnvelope) ===
            "SessionEndedRequest"
        );
    },
    handle(handlerInput) {
        // Any cleanup logic goes here.
        return handlerInput.responseBuilder.getResponse();
    }
};

// The intent reflector is used for interaction model testing and debugging.
// It will simply repeat the intent the user said. You can create custom handlers
// for your intents by defining them above, then also adding them to the request
// handler chain below.
const IntentReflectorHandler = {
    canHandle(handlerInput) {
        return (
            Alexa.getRequestType(handlerInput.requestEnvelope) ===
            "IntentRequest"
        );
    },
    handle(handlerInput) {
        const intentName = Alexa.getIntentName(handlerInput.requestEnvelope);
        const speakOutput = `Hai lanciato ${intentName}`;

        return (
            handlerInput.responseBuilder
                .speak(speakOutput)
                //.reprompt('add a reprompt if you want to keep the session open for the user to respond')
                .getResponse()
        );
    }
};

// Generic error handling to capture any syntax or routing errors. If you receive an error
// stating the request handler chain is not found, you have not implemented a handler for
// the intent being invoked or included it in the skill builder below.
const ErrorHandler = {
    canHandle() {
        return true;
    },
    handle(handlerInput, error) {
        console.log(`~~~~ Error handled: ${error.stack}`);
        const speakOutput = `Non ho capito cosa mi hai chiesto`;

        return handlerInput.responseBuilder
            .speak(speakOutput)
            .reprompt(speakOutput)
            .getResponse();
    }
};

// The SkillBuilder acts as the entry point for your skill, routing all request and response
// payloads to the handlers above. Make sure any new handlers or interceptors you've
// defined are included below. The order matters - they're processed top to bottom.
exports.handler = Alexa.SkillBuilders.custom()
    .addRequestHandlers(
        LaunchRequestHandler,
        SwitchOffTheLightHandler,
        SwitchOnTheLightHandler,
        HelpIntentHandler,
        CancelAndStopIntentHandler,
        SessionEndedRequestHandler,
        IntentReflectorHandler // make sure IntentReflectorHandler is last so it doesn't override your custom intent handlers
    )
    .addErrorHandlers(ErrorHandler)
    .lambda();
