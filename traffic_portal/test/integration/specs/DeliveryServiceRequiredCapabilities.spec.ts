/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */
import { browser } from 'protractor';
import { API } from '../CommonUtils/API';
import { LoginPage } from '../PageObjects/LoginPage.po';
import { TopNavigationPage } from '../PageObjects/TopNavigationPage.po';
import { DeliveryServicesPage } from '../PageObjects/DeliveryServicesPage.po';

let fs = require('fs')
let using = require('jasmine-data-provider');

let testFile = 'Data/DeliveryServiceRequiredCapabilities/TestCases.json';
let setupFile = 'Data/DeliveryServiceRequiredCapabilities/Setup.json';
let cleanupFile = 'Data/DeliveryServiceRequiredCapabilities/Cleanup.json';

let testData = JSON.parse(fs.readFileSync(testFile));

let api = new API();
let loginPage = new LoginPage();
let topNavigation = new TopNavigationPage();
let deliveryServicesPage = new DeliveryServicesPage();

using(testData.DeliveryServiceRequiredCapabilities, async function(data) {
    using(data.Login, function(login) {
        describe('Traffic Portal - Delivery Service Required Capabilities - ' + login.description,  function(){
            it('Setup', async function() {
                let setupData = JSON.parse(fs.readFileSync(setupFile));
                let output = await api.UseAPI(setupData);
                expect(output).toBeNull();
            })
            it(login.description, async function(){
                browser.get(browser.params.baseUrl);
                await loginPage.Login(login.username, login.password);
                expect(await loginPage.CheckUserName(login.username)).toBeTruthy();
            })
            it('can open delivery service page', async function(){
                await deliveryServicesPage.OpenServicesMenu();
                await deliveryServicesPage.OpenDeliveryServicePage();
            })
            using(data.Add, function(test){
                it(test.description, async function(){
                    await deliveryServicesPage.SearchDeliveryService(test.DeliveryService);
                    expect(await deliveryServicesPage.AddRequiredServerCapabilities(test.ServerCapability, test.validationMessage)).toBeTruthy();
                    await deliveryServicesPage.OpenDeliveryServicePage();
                })
            })
            using(data.Remove, function(test){
                it(test.description, async function(){
                    await deliveryServicesPage.SearchDeliveryService(test.DeliveryService);
                    expect(await deliveryServicesPage.RemoveRequiredServerCapabilities(test.ServerCapability, test.validationMessage)).toBeTruthy();
                })
            })
            it('can logout', async function(){
                expect(await topNavigation.Logout()).toBeTruthy();
            })
            it('Cleanup', async function() {
                let cleanupData = JSON.parse(fs.readFileSync(cleanupFile));
                let output = await api.UseAPI(cleanupData);
                expect(output).toBeNull();
            })
        })
    })
})