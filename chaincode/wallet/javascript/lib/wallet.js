'use strict';

const {Contract} = require('fabric-contract-api');

class Wallet extends Contract {

    async initLedger(ctx) {
        console.info('============= START : Initialize Ledger ===========');
        const records = [
            {
                from_pos: '李四',
                to_pos: '赵五',
                amount: '10',
                time: '1548737871',
            },
            {
                from_pos: '赵五',
                to_pos: '李四',
                amount: '30',
                time: '1548757871',
            },
        ];

        for (let i = 0; i < records.length; i++) {
            records[i].docType = 'ledger';
            await ctx.stub.putState(Buffer.from(JSON.stringify(records[i])));
            console.info('Added <--> ', records[i]);
        }
        console.info('============= END : Initialize Ledger ===========');
    }

    async queryTransferRecordByFrom(ctx, from_pos) {
        const recordAsBytes = await ctx.stub.getState(from_pos); // get the record from chaincode state
        if (!recordAsBytes || recordAsBytes.length === 0) {
            throw new Error(`${from_pos} does not exist`);
        }
        console.log(recordAsBytes.toString());
        return recordAsBytes.toString();
    }

    async queryTransferRecordByTo(ctx, to_pos) {
        const recordAsBytes = await ctx.stub.getState(to_pos); // get the record from chaincode state
        if (!recordAsBytes || recordAsBytes.length === 0) {
            throw new Error(`${to_pos} does not exist`);
        }
        console.log(recordAsBytes.toString());
        return recordAsBytes.toString();
    }

    async createTransferRecord(ctx, from_pos, to_pos, amount, transfer_time) {
        console.info('============= START : Create Transfer Record ===========');

        const transfer_record = {
            docType: 'ledger',
            from_pos,
            to_pos,
            amount,
            transfer_time,
        };

        await ctx.stub.putState(Buffer.from(JSON.stringify(transfer_record)));
        console.info('============= END : Create Transfer Record ===========');
    }

}

module.exports = Wallet;
