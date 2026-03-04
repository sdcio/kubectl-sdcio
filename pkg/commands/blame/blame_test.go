package blame_test

import (
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
)

// BuildTestBlameTree returns a complex BlameTreeElement for testing purposes
func BuildTestBlameTree() *sdcpb.BlameTreeElement {
	return &sdcpb.BlameTreeElement{
		Name: "root",
		Childs: []*sdcpb.BlameTreeElement{
			// choices
			{
				Name: "choices",
				Childs: []*sdcpb.BlameTreeElement{
					{
						Name: "case1",
						Childs: []*sdcpb.BlameTreeElement{
							{
								Name: "case-elem",
								Childs: []*sdcpb.BlameTreeElement{
									{
										Name:  "elem",
										Owner: "owner1",
										Value: &sdcpb.TypedValue{
											Value: &sdcpb.TypedValue_StringVal{
												StringVal: "foocaseval",
											},
										},
									},
								},
							},
							{
								Name:  "log",
								Owner: "default",
								Value: &sdcpb.TypedValue{
									Value: &sdcpb.TypedValue_StringVal{
										StringVal: "false",
									},
								},
							},
						},
					},
				},
			},
			// interface
			{
				Name: "interface",
				Childs: []*sdcpb.BlameTreeElement{
					// ethernet-1/1
					{
						Name:    "ethernet-1/1",
						KeyName: "name",
						Childs: []*sdcpb.BlameTreeElement{
							{
								Name:  "admin-state",
								Owner: "owner2",
								Value: &sdcpb.TypedValue{
									Value: &sdcpb.TypedValue_StringVal{
										StringVal: "enable",
									},
								},
							},
							{
								Name:  "description",
								Owner: "owner1",
								Value: &sdcpb.TypedValue{
									Value: &sdcpb.TypedValue_StringVal{
										StringVal: "Changed Description",
									},
								},
								DeviationValue: &sdcpb.TypedValue{
									Value: &sdcpb.TypedValue_StringVal{
										StringVal: "Foo",
									},
								},
							},
							{
								Name:  "name",
								Owner: "owner2",
								Value: &sdcpb.TypedValue{
									Value: &sdcpb.TypedValue_StringVal{
										StringVal: "ethernet-1/1",
									},
								},
							},
							{
								Name: "subinterface",
								Childs: []*sdcpb.BlameTreeElement{
									{
										Name:    "0",
										KeyName: "index",
										Childs: []*sdcpb.BlameTreeElement{
											{
												Name:  "admin-state",
												Owner: "default",
												Value: &sdcpb.TypedValue{
													Value: &sdcpb.TypedValue_StringVal{
														StringVal: "enable",
													},
												},
											},
											{
												Name:  "description",
												Owner: "owner3",
												Value: &sdcpb.TypedValue{
													Value: &sdcpb.TypedValue_StringVal{
														StringVal: "Subinterface 0",
													},
												},
											},
											{
												Name:  "index",
												Owner: "owner3",
												Value: &sdcpb.TypedValue{
													Value: &sdcpb.TypedValue_StringVal{
														StringVal: "0",
													},
												},
											},
											{
												Name:  "type",
												Owner: "owner4",
												Value: &sdcpb.TypedValue{
													Value: &sdcpb.TypedValue_StringVal{
														StringVal: "routed",
													},
												},
											},
										},
									},
								},
							},
						},
					},
					// ethernet-1/2
					{
						Name:    "ethernet-1/2",
						KeyName: "name",
						Childs: []*sdcpb.BlameTreeElement{
							{
								Name:  "admin-state",
								Owner: "owner4",
								Value: &sdcpb.TypedValue{
									Value: &sdcpb.TypedValue_StringVal{
										StringVal: "enable",
									},
								},
							},
							{
								Name:  "description",
								Owner: "owner4",
								Value: &sdcpb.TypedValue{
									Value: &sdcpb.TypedValue_StringVal{
										StringVal: "Foo",
									},
								},
							},
							{
								Name:  "name",
								Owner: "owner2",
								Value: &sdcpb.TypedValue{
									Value: &sdcpb.TypedValue_StringVal{
										StringVal: "ethernet-1/2",
									},
								},
							},
							{
								Name: "subinterface",
								Childs: []*sdcpb.BlameTreeElement{
									{
										Name:    "5",
										KeyName: "index",
										Childs: []*sdcpb.BlameTreeElement{
											{
												Name:  "admin-state",
												Owner: "default",
												Value: &sdcpb.TypedValue{
													Value: &sdcpb.TypedValue_StringVal{
														StringVal: "enable",
													},
												},
											},
											{
												Name:  "description",
												Owner: "owner2",
												Value: &sdcpb.TypedValue{
													Value: &sdcpb.TypedValue_StringVal{
														StringVal: "Subinterface 5",
													},
												},
											},
											{
												Name:  "index",
												Owner: "owner2",
												Value: &sdcpb.TypedValue{
													Value: &sdcpb.TypedValue_StringVal{
														StringVal: "5",
													},
												},
											},
											{
												Name:  "type",
												Owner: "owner2",
												Value: &sdcpb.TypedValue{
													Value: &sdcpb.TypedValue_StringVal{
														StringVal: "routed",
													},
												},
											},
										},
									},
								},
							},
						},
					},
					// ethernet-1/3
					{
						Name:    "ethernet-1/3",
						KeyName: "name",
						Childs: []*sdcpb.BlameTreeElement{
							{
								Name:  "admin-state",
								Owner: "default",
								Value: &sdcpb.TypedValue{
									Value: &sdcpb.TypedValue_StringVal{
										StringVal: "enable",
									},
								},
							},
							{
								Name:  "description",
								Owner: "running",
								Value: &sdcpb.TypedValue{
									Value: &sdcpb.TypedValue_StringVal{
										StringVal: "ethernet-1/3 description",
									},
								},
							},
							{
								Name:  "name",
								Owner: "running",
								Value: &sdcpb.TypedValue{
									Value: &sdcpb.TypedValue_StringVal{
										StringVal: "ethernet-1/3",
									},
								},
							},
						},
					},
				},
			},
			// leaflist
			{
				Name: "leaflist",
				Childs: []*sdcpb.BlameTreeElement{
					{
						Name:  "entry",
						Owner: "owner1",
						Value: &sdcpb.TypedValue{
							Value: &sdcpb.TypedValue_LeaflistVal{
								LeaflistVal: &sdcpb.ScalarArray{
									Element: []*sdcpb.TypedValue{
										{Value: &sdcpb.TypedValue_StringVal{StringVal: "foo"}},
										{Value: &sdcpb.TypedValue_StringVal{StringVal: "bar"}},
									},
								},
							},
						},
					},
					{
						Name:  "with-default",
						Owner: "default",
						Value: &sdcpb.TypedValue{
							Value: &sdcpb.TypedValue_LeaflistVal{
								LeaflistVal: &sdcpb.ScalarArray{
									Element: []*sdcpb.TypedValue{
										{Value: &sdcpb.TypedValue_StringVal{StringVal: "foo"}},
										{Value: &sdcpb.TypedValue_StringVal{StringVal: "bar"}},
									},
								},
							},
						},
					},
				},
			},
			// network-instance
			{
				Name: "network-instance",
				Childs: []*sdcpb.BlameTreeElement{
					// default
					{
						Name:    "default",
						KeyName: "name",
						Childs: []*sdcpb.BlameTreeElement{
							{
								Name:  "admin-state",
								Owner: "owner1",
								Value: &sdcpb.TypedValue{
									Value: &sdcpb.TypedValue_StringVal{
										StringVal: "disable",
									},
								},
							},
							{
								Name:  "description",
								Owner: "owner5",
								Value: &sdcpb.TypedValue{
									Value: &sdcpb.TypedValue_StringVal{
										StringVal: "Default NI",
									},
								},
							},
							{
								Name:  "name",
								Owner: "owner1",
								Value: &sdcpb.TypedValue{
									Value: &sdcpb.TypedValue_StringVal{
										StringVal: "default",
									},
								},
							},
							{
								Name:  "type",
								Owner: "owner5",
								Value: &sdcpb.TypedValue{
									Value: &sdcpb.TypedValue_StringVal{
										StringVal: "default",
									},
								},
							},
						},
					},
					// other
					{
						Name:    "other",
						KeyName: "name",
						Childs: []*sdcpb.BlameTreeElement{
							{
								Name:  "admin-state",
								Owner: "owner2",
								Value: &sdcpb.TypedValue{
									Value: &sdcpb.TypedValue_StringVal{
										StringVal: "enable",
									},
								},
							},
							{
								Name:  "description",
								Owner: "owner2",
								Value: &sdcpb.TypedValue{
									Value: &sdcpb.TypedValue_StringVal{
										StringVal: "Other NI",
									},
								},
							},
							{
								Name:  "name",
								Owner: "owner2",
								Value: &sdcpb.TypedValue{
									Value: &sdcpb.TypedValue_StringVal{
										StringVal: "other",
									},
								},
							},
							{
								Name:  "type",
								Owner: "owner2",
								Value: &sdcpb.TypedValue{
									Value: &sdcpb.TypedValue_StringVal{
										StringVal: "ip-vrf",
									},
								},
							},
						},
					},
				},
			},
			// patterntest
			{
				Name:  "patterntest",
				Owner: "owner1",
				Value: &sdcpb.TypedValue{
					Value: &sdcpb.TypedValue_StringVal{
						StringVal: "hallo 0",
					},
				},
				DeviationValue: &sdcpb.TypedValue{
					Value: &sdcpb.TypedValue_StringVal{
						StringVal: "hallo 00",
					},
				},
			},
		},
	}
}
